package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetReturns(c *gin.Context) {
	rows, err := h.DB.Query(`
		SELECT r.id, r.quantity, r.return_type, COALESCE(r.reason, ''), r.status, 
		       COALESCE(r.reception_id, 0), r.created_at,
		       COALESCE(wp.name, 'Unknown Product') as product_name,
		       COALESCE(rec.category, 'Unknown Supplier') as supplier
		FROM returns r
		LEFT JOIN warehouse_product wp ON r.product_id = wp.id
		LEFT JOIN receptions rec ON r.reception_id = rec.id
		ORDER BY r.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var returns []map[string]interface{}
	for rows.Next() {
		var id, quantity, receptionID int
		var returnType, reason, status, createdAt, productName, supplier string
		
		err := rows.Scan(&id, &quantity, &returnType, &reason, &status, &receptionID, &createdAt, &productName, &supplier)
		if err != nil {
			continue
		}
		
		returns = append(returns, map[string]interface{}{
			"id": id,
			"product_name": productName,
			"quantity": quantity,
			"return_type": returnType,
			"reason": reason,
			"status": status,
			"supplier": supplier,
			"reception_id": receptionID,
			"created_at": createdAt[:10],
		})
	}

	c.JSON(http.StatusOK, returns)
}

func (h *Handler) CreateReturn(c *gin.Context) {
	var req struct {
		ProductName string `json:"product_name"`
		Quantity    int    `json:"quantity"`
		ReturnType  string `json:"return_type"`
		Reason      string `json:"reason"`
		Supplier    string `json:"supplier"`
		ReceptionID int    `json:"reception_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	err := h.DB.QueryRow(`
		INSERT INTO returns (product_name, quantity, return_type, reason, supplier, reception_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, req.ProductName, req.Quantity, req.ReturnType, req.Reason, req.Supplier, req.ReceptionID).Scan(&id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Return created successfully"})
}

func (h *Handler) UpdateReturnStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid return ID"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = h.DB.Exec(`
		UPDATE returns 
		SET status = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`, req.Status, id)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Return status updated successfully"})
}
