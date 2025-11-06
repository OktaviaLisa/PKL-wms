package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetInventoryData(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT i.id, i.product_name, COALESCE(i.category, 'Unknown') as category, 
		       i.quantity, i.location_id, i.updated_at,
		       COALESCE(l.name, 'Unknown Location') as location
		FROM inventory i
		LEFT JOIN locations l ON i.location_id = l.id
		WHERE i.quantity > 0 
		ORDER BY i.product_name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventory: " + err.Error()})
		return
	}
	defer rows.Close()

	var inventory []map[string]interface{}
	for rows.Next() {
		var id, quantity, locationID int
		var productName, category, updatedAt, location string

		if err := rows.Scan(&id, &productName, &category, &quantity, &locationID, &updatedAt, &location); err != nil {
			continue
		}

		inventory = append(inventory, map[string]interface{}{
			"id":           id,
			"product_name": productName,
			"category":     category,
			"quantity":     quantity,
			"location_id":  locationID,
			"location":     location,
			"updated_at":   updatedAt[:19], // Format datetime
		})
	}

	c.JSON(http.StatusOK, inventory)
}

func (h *Handler) CreateInventoryItem(c *gin.Context) {
	var item struct {
		ProductName string `json:"product_name"`
		Category    string `json:"category"`
		Quantity    int    `json:"quantity"`
		Location    string `json:"location"`
		MinStock    int    `json:"min_stock"`
	}

	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Check if item already exists
	var existingID int
	err := h.DB.QueryRow(
		"SELECT id FROM inventory WHERE product_name = $1 AND category = $2",
		item.ProductName, item.Category,
	).Scan(&existingID)

	if err == nil {
		// Item exists, update quantity
		_, err = h.DB.Exec(
			"UPDATE inventory SET quantity = quantity + $1, updated_at = NOW() WHERE id = $2",
			item.Quantity, existingID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Inventory updated", "id": existingID})
	} else {
		// Item doesn't exist, create new
		var newID int
		err = h.DB.QueryRow(
			"INSERT INTO inventory (product_name, category, quantity, location, min_stock, updated_at) VALUES ($1, $2, $3, $4, $5, NOW()) RETURNING id",
			item.ProductName, item.Category, item.Quantity, item.Location, item.MinStock,
		).Scan(&newID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create inventory item"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Inventory item created", "id": newID})
	}
}
