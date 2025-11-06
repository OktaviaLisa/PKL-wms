package handlers

import (
	"net/http"
	"time"
	"wms-backend/internal/database"
	"wms-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Get all quality checks
func GetQualityCheck(c *gin.Context) {
	rows, err := database.DB.Query(`SELECT id, reception_id, product_name, quantity, status, notes, checked_at FROM quality_checks`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var checks []models.QualityCheck
	for rows.Next() {
		var qc models.QualityCheck
		if err := rows.Scan(&qc.ID, &qc.ReceptionID, &qc.ProductName, &qc.Quantity, &qc.Status, &qc.Notes, &qc.CheckedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		checks = append(checks, qc)
	}

	c.JSON(http.StatusOK, checks)
}

// Create new quality check
// Create new quality check
func CreateQualityCheck(c *gin.Context) {
	var input models.QualityCheck
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.CheckedAt = time.Now()

	// Simpan QC ke tabel quality_checks
	_, err := database.DB.Exec(`
		INSERT INTO quality_checks (reception_id, product_name, quantity, status, notes, checked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, input.ReceptionID, input.ProductName, input.Quantity, input.Status, input.Notes, input.CheckedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Kalau status FAIL → buat entri di tabel returns
	if input.Status == "FAIL" {
		_, err = database.DB.Exec(`
			INSERT INTO returns (quality_check_id, reception_id, product_name, quantity, reason, return_type, status, created_at)
			VALUES (
				(SELECT MAX(id) FROM quality_checks),
				$1, $2, $3, $4, 'QUALITY_FAIL', 'PENDING', NOW()
			)
		`, input.ReceptionID, input.ProductName, input.Quantity, input.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Quality Check saved, but failed to create return: " + err.Error()})
			return
		}
	}

	// Kalau status PASS → update / insert ke tabel inventory
	if input.Status == "PASS" {
		// Cari atau buat product_id
		var productID int
		err = database.DB.QueryRow(`
			SELECT id FROM warehouse_product WHERE name = $1
		`, input.ProductName).Scan(&productID)
		
		if err != nil {
			// Product belum ada, buat baru
			err = database.DB.QueryRow(`
				INSERT INTO warehouse_product (name, sku, category_id, description, price, created_at)
				VALUES ($1, $2, 1, 'Auto-created from QC', 0, NOW())
				RETURNING id
			`, input.ProductName, "AUTO-"+input.ProductName).Scan(&productID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
				return
			}
		}

		// Cek apakah sudah ada di inventory
		var count int
		err = database.DB.QueryRow(`
			SELECT COUNT(*) FROM inventory WHERE product_id = $1
		`, productID).Scan(&count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check inventory: " + err.Error()})
			return
		}

		if count == 0 {
			// Belum ada → buat data baru
			_, err = database.DB.Exec(`
				INSERT INTO inventory (product_id, product_name, quantity, min_stock, location_id, updated_at)
				VALUES ($1, $2, $3, 0, 1, NOW())
			`, productID, input.ProductName, input.Quantity)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert inventory: " + err.Error()})
				return
			}
		} else {
			// Sudah ada → update quantity
			_, err = database.DB.Exec(`
				UPDATE inventory
				SET quantity = quantity + $1, updated_at = NOW()
				WHERE product_id = $2
			`, input.Quantity, productID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory: " + err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Quality check created successfully"})
}
