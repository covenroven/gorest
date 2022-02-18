package api

import (
	"database/sql"
	"net/http"

	"github.com/covenroven/gorest/model"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// Get all orders
func (a *App) GetOrders(c *gin.Context) {
	query := `
		SELECT order_id, customer_name, ordered_at
		FROM orders
		ORDER BY order_id ASC;
	`
	rows, err := a.DB.Query(query)
	if err != nil {
		throwServerError(c, err)
		return
	}
	defer rows.Close()

	var orders []model.Order
	orderIDs := []uint{}
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.OrderID,
			&order.CustomerName,
			&order.OrderedAt,
		); err != nil {
			throwServerError(c, err)
			return
		}

		orders = append(orders, order)
		orderIDs = append(orderIDs, order.OrderID)
	}

	// Fetch items
	if len(orders) > 0 {
		query = `
			SELECT item_id, item_code, description, quantity, order_id
			FROM items
			WHERE order_id = ANY($1)
			ORDER BY order_id ASC;
		`
		rows, err = a.DB.Query(query, pq.Array(orderIDs))
		if err != nil {
			throwServerError(c, err)
			return
		}
		defer rows.Close()

		items := map[uint][]model.Item{}
		for rows.Next() {
			var item model.Item
			if err := rows.Scan(
				&item.ItemID,
				&item.ItemCode,
				&item.Description,
				&item.Quantity,
				&item.OrderID,
			); err != nil {
				throwServerError(c, err)
				return
			}

			items[item.OrderID] = append(items[item.OrderID], item)
		}

		for i := range orders {
			orders[i].Items = items[orders[i].OrderID]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": orders,
	})
}

// Show an order by ID
func (a *App) ShowOrder(c *gin.Context) {
	orderID := c.Param("orderID")
	order, err := a.findOrder(orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"data": nil,
			})
			return
		} else {
			throwServerError(c, err)
		}
	}

	// Fetch items
	query := `
		SELECT item_id, item_code, description, quantity, order_id
		FROM items
		WHERE order_id = $1
		ORDER BY order_id ASC;
	`
	rows, err := a.DB.Query(query, orderID)
	if err != nil {
		throwServerError(c, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item model.Item
		if err := rows.Scan(
			&item.ItemID,
			&item.ItemCode,
			&item.Description,
			&item.Quantity,
			&item.OrderID,
		); err != nil {
			throwServerError(c, err)
			return
		}

		order.Items = append(order.Items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": order,
	})
}

// Create a new order
func (a *App) CreateOrder(c *gin.Context) {
	var order model.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Insert order
	query := `
		INSERT INTO orders (customer_name, ordered_at)
		VALUES ($1, $2)
		RETURNING order_id;
	`
	err := a.DB.QueryRow(query, order.CustomerName, order.OrderedAt).
		Scan(&order.OrderID)
	if err != nil {
		throwServerError(c, err)
		return
	}

	// Insert items
	for i, item := range order.Items {
		query = `
			INSERT INTO items (item_code, description, quantity, order_id)
			VALUES ($1, $2, $3, $4)
			RETURNING item_id;
		`
		err := a.DB.QueryRow(query, item.ItemCode, item.Description, item.Quantity, order.OrderID).
			Scan(&order.Items[i].ItemID)
		if err != nil {
			throwServerError(c, err)
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": order,
	})
}

// Update an order by ID
func (a *App) UpdateOrder(c *gin.Context) {
	orderID := c.Param("orderID")
	_, err := a.findOrder(orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"data": nil,
			})
			return
		} else {
			throwServerError(c, err)
			return
		}
	}

	var orderRequest model.Order
	if err := c.ShouldBindJSON(&orderRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	query := `
		UPDATE orders
		SET customer_name = $1, ordered_at = $2
		WHERE order_id = $3
		RETURNING order_id;
	`
	err = a.DB.QueryRow(query, orderRequest.CustomerName, orderRequest.OrderedAt, orderID).
		Scan(&orderRequest.OrderID)
	if err != nil {
		throwServerError(c, err)
		return
	}

	// Update items
	var updatedItems []model.Item
	for i, item := range orderRequest.Items {
		query = `
			UPDATE items 
			SET item_code = $1, description = $2, quantity = $3
			WHERE item_id = $4 AND order_id = $5;
		`
		res, err := a.DB.Exec(query, item.ItemCode, item.Description, item.Quantity, item.ItemID, orderID)
		if err != nil {
			throwServerError(c, err)
			return
		}
		count, _ := res.RowsAffected()
		if count > 0 {
			orderRequest.Items[i].OrderID = orderRequest.OrderID
			updatedItems = append(updatedItems, orderRequest.Items[i])
		}
	}
	orderRequest.Items = updatedItems

	c.JSON(http.StatusOK, gin.H{
		"data": orderRequest,
	})
}

// Delete an order by ID
func (a *App) DeleteOrder(c *gin.Context) {
	orderID := c.Param("orderID")
	_, err := a.findOrder(orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"data": nil,
			})
			return
		} else {
			throwServerError(c, err)
			return
		}
	}

	// Delete items first
	query := `
		DELETE FROM items
		WHERE order_id = $1
	`
	_, err = a.DB.Exec(query, orderID)
	if err != nil {
		throwServerError(c, err)
		return
	}

	// Delete order
	query = `
		DELETE FROM orders
		WHERE order_id = $1
	`
	_, err = a.DB.Exec(query, orderID)
	if err != nil {
		throwServerError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

// Fetch Order by ID
func (a *App) findOrder(orderID string) (model.Order, error) {
	query := `
		SELECT o.order_id, o.customer_name, o.ordered_at
		FROM orders o
		WHERE o.order_id = $1;
	`
	row := a.DB.QueryRow(query, orderID)

	var order model.Order
	if err := row.Scan(
		&order.OrderID,
		&order.CustomerName,
		&order.OrderedAt,
	); err != nil {
		return order, err
	}

	return order, nil
}
