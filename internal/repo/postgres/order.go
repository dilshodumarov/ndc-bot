package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/pkg/postgres"

	"github.com/jackc/pgx"
)

// OrderRepo -.
type OrderRepo struct {
	*postgres.Postgres
}

// New -.
func NewOrderRepo(pg *postgres.Postgres) *OrderRepo {
	return &OrderRepo{pg}
}

// func (r *OrderRepo) CreateOrder(ctx context.Context, order entity.Order) (*entity.CreateOrderResponse, error) {
// 	tx, err := r.Pool.Begin(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("OrderRepo - CreateOrder - Begin: %w", err)
// 	}
// 	defer tx.Rollback(ctx)
// 	platformID := fmt.Sprintf("%d", order.UserID)
// 	var clientID, Username sql.NullString
// 	err = tx.QueryRow(ctx, `
// 		SELECT c.guid, i.integration_user_name
// 		FROM integration AS i
// 		LEFT JOIN client AS c ON c.bussnes_id = i.owner_id AND c.platform_id = $1
// 		WHERE i.owner_id = $2
// 	`, platformID, order.BusinessId).Scan(&clientID, &Username)

// 	if err != nil {

// 		return nil, err
// 	}

// 	if !clientID.Valid {
// 		return &entity.CreateOrderResponse{Message: "norows", TgUserName: Username.String}, nil
// 	}

// 	switch order.Status {
// 	// case "tolov_qilishga_tayyor":
// 	// 	_, err = tx.Exec(ctx, `
// 	// 		DELETE FROM "order"
// 	// 		WHERE status = 'yangi'
// 	// 		AND client_id = $1
// 	// 	`, clientID)
// 	// 	if err != nil {
// 	// 		log.Println(222222222, err)
// 	// 		return nil, fmt.Errorf("CreateOrder - Delete old orders (yangi): %w", err)
// 	// 	}
// 	case "tolov_qilmoqchi":
// 		_, err = tx.Exec(ctx, `
// 			DELETE FROM "order"
// 			WHERE status = 'yangi'
// 			AND client_id = $1
// 		`, clientID)
// 		if err != nil {

// 			return nil, fmt.Errorf("CreateOrder - Delete old orders (yangi): %w", err)
// 		}
// 	}

// 	var orderID string
// 	var OrderIdSerial int64
// 	err = tx.QueryRow(ctx, `
// 		INSERT INTO "order" (client_id, status, status_changed_time, status_number,platform,business_id, order_status_id,location_url, total_price,created_at, updated_at)
// 		VALUES ($1, $2, CURRENT_TIMESTAMP, $3, $4, $5,$6,$7,$8,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
// 		RETURNING guid,order_id
// 	`, clientID.String, order.Status, order.StatusNumber, order.Platform, order.BusinessId, order.StatusId, order.Location, order.TotalPrice).Scan(&orderID, &OrderIdSerial)

// 	if err != nil {

// 		return nil, fmt.Errorf("OrderRepo - CreateOrder - QueryRow (order): %w", err)
// 	}

// 	for _, pid := range order.ProductOrders {
// 		_, err = tx.Exec(ctx, `
// 			INSERT INTO order_products (order_id, product_id, count, price, total_price,created_at, updated_at)
// 			VALUES ($1, $2, $3, $4, $5,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
// 		`, orderID, pid.Id, pid.Count, pid.Price, pid.TotalPrice)
// 		if err != nil {
// 			log.Println("INSERT order_products error:", err)
// 			return nil, fmt.Errorf("OrderRepo - CreateOrder - Insert (order_products): %w", err)
// 		}

// 		if order.Status == "buyurtma_qilmoqchi" {
// 			_, err = tx.Exec(ctx, `
// 				UPDATE product
// 				SET count = count - $1,
// 					updated_at = CURRENT_TIMESTAMP
// 				WHERE guid = $2 AND business_id = $3 AND count >= $1
// 			`, pid.Count, pid.Id, order.BusinessId)
// 			if err != nil {
// 				log.Println("UPDATE product count error:", err)
// 				return nil, fmt.Errorf("OrderRepo - CreateOrder - Update (product count): %w", err)
// 			}
// 		}
// 	}

// 	var TgUSerName string
// 	err = tx.QueryRow(ctx, `
// 		SELECT tg_user_name FROM business WHERE guid = $1
// 	`, order.BusinessId).Scan(&TgUSerName)

// 	if err != nil {

// 		return nil, fmt.Errorf("CreateOrder - Select TgUserName: %w", err)
// 	}

// 	if err := tx.Commit(ctx); err != nil {
// 		return nil, fmt.Errorf("OrderRepo - CreateOrder - Commit: %w", err)
// 	}

// 	return &entity.CreateOrderResponse{
// 		Id:            orderID,
// 		TgUserName:    TgUSerName,
// 		OrderIdSerial: OrderIdSerial,
// 	}, nil
// }

func (r *OrderRepo) CreateOrder(ctx context.Context, order entity.Order) (*entity.CreateOrderResponse, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("OrderRepo - CreateOrder - Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	platformID := fmt.Sprintf("%d", order.UserID)
	var clientID, Username sql.NullString
	err = tx.QueryRow(ctx, `
		SELECT c.guid, i.integration_user_name
		FROM integration AS i
		LEFT JOIN client AS c ON c.bussnes_id = i.owner_id AND c.platform_id = $1
		WHERE i.owner_id = $2
	`, platformID, order.BusinessId).Scan(&clientID, &Username)
	if err != nil {
		return nil, err
	}
	if !clientID.Valid {
		return &entity.CreateOrderResponse{Message: "norows", TgUserName: Username.String}, nil
	}

	var orderID string
	var OrderIdSerial int64

	// Agar status = "yangi" bo‘lsa eski yangi statusdagi order borligini tekshiramiz
	if order.Status == "yangi" {
		err = tx.QueryRow(ctx, `
			SELECT guid, order_id FROM "order"
			WHERE client_id = $1 AND status = 'yangi' AND deleted_at IS NULL
			LIMIT 1
		`, clientID.String).Scan(&orderID, &OrderIdSerial)

		if err != nil {
			if err == pgx.ErrNoRows || err.Error() == "no rows in result set" {

				// eski topilmadi - yangi yaratamiz
				err = tx.QueryRow(ctx, `
					INSERT INTO "order" (client_id, status, status_changed_time, status_number, platform, business_id, order_status_id, location_url, total_price, created_at, updated_at)
					VALUES ($1, $2, CURRENT_TIMESTAMP, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
					RETURNING guid, order_id
				`, clientID.String, order.Status, order.StatusNumber, order.Platform, order.BusinessId, order.StatusId, order.Location, order.TotalPrice).Scan(&orderID, &OrderIdSerial)

				if err != nil {
					return nil, fmt.Errorf("OrderRepo - CreateOrder - Insert new (yangi): %w", err)
				}
			} else {

				return nil, fmt.Errorf("OrderRepo - CreateOrder - Select existing (yangi): %w", err)
			}
		}

	} else {

		var deletedOrderID sql.NullString
		err = tx.QueryRow(ctx, `
			UPDATE "order"
			SET deleted_at = CURRENT_TIMESTAMP 
			WHERE client_id = $1 AND status = 'yangi' AND deleted_at IS NULL
			RETURNING guid
		`, clientID.String).Scan(&deletedOrderID)
		if err != nil {
			return nil, fmt.Errorf("OrderRepo - CreateOrder - Soft delete old yangi: %w", err)
		}

		err = tx.QueryRow(ctx, `
		INSERT INTO "order" (
			client_id, order_guid, status, status_changed_time, status_number,
			platform, business_id, order_status_id, location_url, total_price,
			created_at, updated_at
		)
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	RETURNING guid, order_id
`,
			clientID.String, deletedOrderID.String, order.Status, order.StatusNumber,
			order.Platform, order.BusinessId, order.StatusId, order.Location, order.TotalPrice,
		).Scan(&orderID, &OrderIdSerial)

		if err != nil {
			return nil, fmt.Errorf("OrderRepo - CreateOrder - Insert (non-yangi): %w", err)
		}
	}

	// Har doim productlar yangiga yoziladi
	for _, pid := range order.ProductOrders {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_products (order_id, product_id, count, price, total_price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, orderID, pid.Id, pid.Count, pid.Price, pid.TotalPrice)
		if err != nil {
			log.Println("INSERT order_products error:", err)
			return nil, fmt.Errorf("OrderRepo - CreateOrder - Insert (order_products): %w", err)
		}

		// Agar buyurtma real bo‘lsa, mahsulotdan ayiramiz
		if order.Status == "buyurtma_qilmoqchi" {
			_, err = tx.Exec(ctx, `
				UPDATE product
				SET count = count - $1, updated_at = CURRENT_TIMESTAMP
				WHERE guid = $2 AND business_id = $3 AND count >= $1
			`, pid.Count, pid.Id, order.BusinessId)
			if err != nil {
				log.Println("UPDATE product count error:", err)
				return nil, fmt.Errorf("OrderRepo - CreateOrder - Update (product count): %w", err)
			}
		}
	}

	var TgUserName string
	err = tx.QueryRow(ctx, `
		SELECT tg_user_name FROM business WHERE guid = $1
	`, order.BusinessId).Scan(&TgUserName)
	if err != nil {
		return nil, fmt.Errorf("CreateOrder - Select TgUserName: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("OrderRepo - CreateOrder - Commit: %w", err)
	}

	return &entity.CreateOrderResponse{
		Id:            orderID,
		TgUserName:    TgUserName,
		OrderIdSerial: OrderIdSerial,
	}, nil
}

func (r *OrderRepo) GetClientOrders(ctx context.Context, platformId, bussnesid string) ([]*entity.OrderResponseByOrderId, error) {
	query := `
		SELECT 
			o.guid AS order_guid,
			o.status,
			o.created_at AS order_created_at,
			p.name AS product_name,
			op.count,
			op.created_at AS product_created_at,
			op.updated_at AS product_updated_at
		FROM client c
		JOIN "order" o ON o.client_id = c.guid
		JOIN order_products op ON op.order_id = o.guid
		JOIN product p ON p.guid = op.product_id
		WHERE c.platform_id = $1 and o.business_id=$2
		ORDER BY o.created_at DESC, op.created_at ASC
	`

	rows, err := r.Pool.Query(ctx, query, platformId, bussnesid)
	if err != nil {
		return nil, fmt.Errorf("OrderRepo - GetClientOrders - Query: %w", err)
	}
	defer rows.Close()

	orderMap := make(map[string]*entity.OrderResponseByOrderId)
	for rows.Next() {
		var (
			orderID        string
			status         string
			orderCreatedAt time.Time
			productName    string
			count          int
			productCreated time.Time
			productUpdated time.Time
		)

		err := rows.Scan(&orderID, &status, &orderCreatedAt, &productName, &count, &productCreated, &productUpdated)
		if err != nil {
			return nil, fmt.Errorf("OrderRepo - GetClientOrders - Scan: %w", err)
		}

		if _, exists := orderMap[orderID]; !exists {
			orderMap[orderID] = &entity.OrderResponseByOrderId{
				Status:    status,
				CreatedAt: orderCreatedAt,
				Products:  []*entity.OrderProduct{},
			}
		}

		orderMap[orderID].Products = append(orderMap[orderID].Products, &entity.OrderProduct{
			ProductName: productName,
			Count:       count,
			CreatedAt:   productCreated,
			UpdatedAt:   productUpdated,
		})
	}

	if len(orderMap) == 0 {
		return nil, nil
	}

	// Convert map to slice
	var orders []*entity.OrderResponseByOrderId
	for _, v := range orderMap {
		orders = append(orders, v)
	}

	return orders, nil
}

func (r *OrderRepo) GetUsersByLastOrder(ctx context.Context) ([]*entity.LastOrders, error) {
	query := `
		SELECT 
			c.chat_id,
			o.business_id
		FROM client c
		JOIN "order" o ON o.client_id = c.guid
		WHERE o.created_at <= NOW() - INTERVAL '1 month'
		ORDER BY o.created_at DESC
	`

	rows, err := r.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("OrderRepo - GetUsersByLastOrder - Query: %w", err)
	}
	defer rows.Close()

	var orders []*entity.LastOrders
	for rows.Next() {
		var order entity.LastOrders
		if err := rows.Scan(&order.ChatId, &order.BotGuid); err != nil {
			return nil, fmt.Errorf("OrderRepo - GetUsersByLastOrder - Scan: %w", err)
		}
		orders = append(orders, &order)
	}

	return orders, nil
}



func (r *OrderRepo) UpdateOrderStatus(ctx context.Context, req *entity.UpdateOrderRequest) (*entity.UpdateOrderResponse, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("UpdateOrderStatus - Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	err = tx.QueryRow(ctx, `
		SELECT status FROM "order" WHERE order_id = $1
	`, req.OrderID).Scan(&currentStatus)
	if err != nil {
		return nil, fmt.Errorf("UpdateOrderStatus - select current status: %w", err)
	}

	if currentStatus == "delivered" {
		return &entity.UpdateOrderResponse{
			Message:  "Buyurtma allaqachon yetkazilgan, holatini o'zgartirib bo'lmaydi.",
			Isupdate: false,
		}, nil
	}
	if currentStatus == "bekor_qilindi" {
		return &entity.UpdateOrderResponse{
			Message:  "Buyurtma allaqachon bekor qilingan, holatini o'zgartirib bo'lmaydi.",
			Isupdate: false,
		}, nil
	}

	setParts := []string{}
	args := []interface{}{}
	argPos := 1

	if req.NewStatus != "" {
		setParts = append(setParts, fmt.Sprintf(`status = $%d`, argPos))
		args = append(args, req.NewStatus)
		argPos++
	}
	if req.Location != "" {
		setParts = append(setParts, fmt.Sprintf(`location = $%d`, argPos))
		args = append(args, req.Location)
		argPos++
	}
	if req.UserNote != "" {
		setParts = append(setParts, fmt.Sprintf(`user_note = $%d`, argPos))
		args = append(args, req.UserNote)
		argPos++
	}
	if req.StatusNumber > 0 {
		setParts = append(setParts, fmt.Sprintf(`status_number = $%d`, argPos))
		args = append(args, req.StatusNumber)
		argPos++
	}
	if req.AdminStatus != nil {
		setParts = append(setParts, fmt.Sprintf(`order_status_id = $%d`, argPos))
		args = append(args, req.AdminStatus)
		argPos++
	}
	if req.LocationUrl != "" {
		setParts = append(setParts, fmt.Sprintf(`location_url = $%d`, argPos))
		args = append(args, req.LocationUrl)
		argPos++
	}
	if req.ImageUrl != "" {
		setParts = append(setParts, fmt.Sprintf(`image_url = $%d`, argPos))
		args = append(args, req.ImageUrl)
		argPos++
	}
	if req.PaymentMethod != "" {
		setParts = append(setParts, fmt.Sprintf(`payment_method = $%d`, argPos))
		args = append(args, req.PaymentMethod)
		argPos++
	}
	setParts = append(setParts, `updated_at = CURRENT_TIMESTAMP`)

	if len(setParts) == 1 { // Faqat updated_at bo‘lsa
		return nil, fmt.Errorf("Yangilanish uchun hech qanday maydon kiritilmadi.")
	}

	query := fmt.Sprintf(`UPDATE "order" SET %s WHERE order_id = $%d`, strings.Join(setParts, ", "), argPos)
	args = append(args, req.OrderID)

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("UpdateOrderStatus - update order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("UpdateOrderStatus - commit: %w", err)
	}

	return &entity.UpdateOrderResponse{
		Message:  "Buyurtma holati muvaffaqiyatli yangilandi.",
		Isupdate: true,
	}, nil
}

func (r *OrderRepo) RestoreProductCounts(ctx context.Context, req entity.CanseledOrder) (*entity.UpdateOrderResponse, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("RestoreProductCounts - begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var orderGUID string
	var currentStatus string
	err = tx.QueryRow(ctx, `
		SELECT guid, status FROM "order"
		WHERE order_id = $1
	`, req.OrderID).Scan(&orderGUID, &currentStatus)
	if err != nil {
		return nil, fmt.Errorf("RestoreProductCounts - select order: %w", err)
	}

	if currentStatus == "delivered" {
		return &entity.UpdateOrderResponse{
			Message:  "Buyurtma allaqachon yetkazib berilgan. Uni bekor qilib bo'lmaydi.",
			Isupdate: false,
		}, nil
	}

	if currentStatus == "bekor_qilindi" {
		return &entity.UpdateOrderResponse{
			Message:  "Buyurtmani allaqachon bekor qilgansiz",
			Isupdate: false,
		}, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT product_id, count
		FROM order_products
		WHERE order_id = $1
	`, orderGUID)
	if err != nil {
		return nil, fmt.Errorf("RestoreProductCounts - select order_products: %w", err)
	}
	defer rows.Close()

	type productUpdate struct {
		productID string
		count     int
	}
	var updates []productUpdate

	for rows.Next() {
		var pu productUpdate
		if err := rows.Scan(&pu.productID, &pu.count); err != nil {
			return nil, fmt.Errorf("RestoreProductCounts - scan order_products: %w", err)
		}
		updates = append(updates, pu)
	}

	for _, pu := range updates {
		_, err := tx.Exec(ctx, `
			UPDATE product
			SET count = count + $1,
				updated_at = CURRENT_TIMESTAMP
			WHERE guid = $2
		`, pu.count, pu.productID)
		if err != nil {
			return nil, fmt.Errorf("RestoreProductCounts - update product count: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE "order"
		SET status = 'bekor_qilindi',
			canceled_reason = $1,
			status_number=$3,
			order_status_id=$4,
			updated_at = CURRENT_TIMESTAMP
		WHERE guid = $2
	`, req.Reason, orderGUID, 6, req.StatusId)
	if err != nil {
		return nil, fmt.Errorf("RestoreProductCounts - update order status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("RestoreProductCounts - commit: %w", err)
	}

	return &entity.UpdateOrderResponse{
		Message:  "Buyurtma muvaffaqiyatli bekor qilindi.",
		Isupdate: true,
	}, nil
}

func (r *OrderRepo) GetOrderByID(ctx context.Context, order entity.GetOrderByID) (*entity.OrderResponseByOrderId, error) {
	id, err := strconv.Atoi(order.OrderID)
	if err != nil {
		return nil, fmt.Errorf("GetOrderByID - invalid orderID: %w", err)
	}

	var clientID sql.NullString
	err = r.Pool.QueryRow(ctx, `
		SELECT c.guid
		FROM client c
		WHERE c.platform_id = $1 AND c.bussnes_id = $2
	`, order.PlatformId, order.BussnesId).Scan(&clientID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("GetOrderByID - QueryRow (client): %w", err)
	}
	if !clientID.Valid {
		return nil, fmt.Errorf("GetOrderByID - clientID is null")
	}

	query := `
		SELECT 
			o.order_id AS order_guid,
			o.status,
			o.status_changed_time,
			o.created_at AS order_created_at,
			p.name AS product_name,
			op.count,
			op.created_at AS product_created_at
		FROM "order" o
		JOIN order_products op ON op.order_id = o.guid
		JOIN product p ON p.guid = op.product_id
		WHERE o.order_id = $1 AND o.client_id = $2 AND o.business_id = $3
		ORDER BY op.created_at ASC
	`

	rows, err := r.Pool.Query(ctx, query, id, clientID.String, order.BussnesId)
	if err != nil {
		return nil, fmt.Errorf("OrderRepo - GetOrderByID - Query: %w", err)
	}
	defer rows.Close()

	orderResponse := &entity.OrderResponseByOrderId{
		Products: []*entity.OrderProduct{},
	}

	for rows.Next() {
		var (
			orderID        int64
			status         string
			statusChanged  time.Time
			orderCreatedAt time.Time
			productName    string
			count          int
			productCreated time.Time
		)

		err := rows.Scan(&orderID, &status, &statusChanged, &orderCreatedAt, &productName, &count, &productCreated)
		if err != nil {
			return nil, fmt.Errorf("OrderRepo - GetOrderByID - Scan: %w", err)
		}

		orderResponse.OrderID = orderID
		orderResponse.Status = status
		orderResponse.StatusChangedTime = statusChanged
		orderResponse.CreatedAt = orderCreatedAt

		orderResponse.Products = append(orderResponse.Products, &entity.OrderProduct{
			ProductName: productName,
			Count:       count,
			CreatedAt:   productCreated,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("OrderRepo - GetOrderByID - Rows error: %w", err)
	}

	return orderResponse, nil
}
