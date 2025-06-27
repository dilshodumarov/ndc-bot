package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/pkg/postgres"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
)

const _defaultEntityCap = 64

// ProductRepo -.
type ProductRepo struct {
	*postgres.Postgres
}

// New -.
func New(pg *postgres.Postgres) *ProductRepo {
	return &ProductRepo{pg}
}


func (r *ProductRepo) GetProduct(ctx context.Context, req entity.GetProductRequest) ([]entity.Product, error) {
	fmt.Println("ProductRepo - GetProduct - name:", req.Name)

	query := `
		SELECT p.guid,p.name, p.image_url,p.cost, p.count, p.guid, p.product_id
		FROM product p
		LEFT JOIN category c ON p.category_id = c.guid
		WHERE (p.name ILIKE $1 
				OR p.description ILIKE $1 
				OR p.short_info ILIKE $1 
				OR c.name ILIKE $1)
		AND p.business_id = $2
	`
	args := []interface{}{"%" + req.Name + "%", req.BusinessID}

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		fmt.Println("❌ Error while getting product:", err)
		return nil, fmt.Errorf("ProductRepo - GetProduct - Query: %w", err)
	}
	defer rows.Close()

	products := make([]entity.Product, 0, 10)

	for rows.Next() {
		var p entity.Product
		//&p.Image_url,
		err := rows.Scan(&p.ID, &p.Name, &p.Cost, &p.Count, &p.ID, &p.ProductId)
		if err != nil {
			return nil, fmt.Errorf("ProductRepo - GetProduct - rows.Scan: %w", err)
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *ProductRepo) GetProductsByAlternatives(ctx context.Context, names []string, businessID string) ([]entity.Product, error) {
	query := r.Builder.
		Select("name, cost, count,product_id").
		From("product")

	// OR bilan barcha ILIKE shartlarini ulaymiz
	orConditions := squirrel.Or{}
	for _, n := range names {
		orConditions = append(orConditions, squirrel.Or{
			squirrel.ILike{"name": "%" + n + "%"},
			squirrel.ILike{"short_info": "%" + n + "%"},
			squirrel.ILike{"description": "%" + n + "%"},
		})
	}

	query = query.Where(
		squirrel.And{
			orConditions,
			squirrel.Eq{"business_id": businessID},
		},
	)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("ProductRepo - GetProductsByAlternatives - ToSql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ProductRepo - GetProductsByAlternatives - Query: %w", err)
	}
	defer rows.Close()

	products := make([]entity.Product, 0, _defaultEntityCap)
	for rows.Next() {
		var p entity.Product
		err := rows.Scan(&p.Name, &p.Cost, &p.Count, &p.ProductId)
		if err != nil {
			return nil, fmt.Errorf("ProductRepo - GetProductsByAlternatives - Scan: %w", err)
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *ProductRepo) GetProductById(ctx context.Context, req entity.GetProductByIDRequest) (*entity.Product, error) {
	var (
		query        string
		args         []interface{}
		picturesting string
	)

	switch {
	case req.BusinessID != "":
		query = `
			SELECT 
				p.guid, 
				p.product_id,
				p.name, 
				p.category_id, 
				p.short_info, 
				p.description,
				p.cost, 
				p.count, 
				p.discount_cost, 
				p.discount, 
				p.created_at, 
				p.updated_at,
				COALESCE(STRING_AGG(pp.image_url, ','), '') AS image_urls
			FROM product p
			LEFT JOIN product_pictures pp ON p.guid = pp.product_id
			WHERE p.product_id = $1 AND p.business_id = $2 AND p.status = true AND p.deleted_at is null
			GROUP BY p.guid
		`
		args = append(args, req.ProductId, req.BusinessID)

	case req.PhoneNumber != "":
		query = `
			SELECT 
				p.guid, 
				p.product_id,
				p.name, 
				p.category_id, 
				p.short_info, 
				p.description,
				p.cost, 
				p.count, 
				p.discount_cost, 
				p.discount, 
				p.created_at, 
				p.updated_at,
				COALESCE(STRING_AGG(pp.image_url, ','), '') AS image_urls
			FROM product p
			JOIN telegram_accaunt t ON p.business_id = t.business_id
			LEFT JOIN product_pictures pp ON p.guid = pp.product_id
			WHERE p.product_id = $1 AND t.number = $2 AND p.status = true AND p.deleted_at is null
			GROUP BY p.guid
		`
		args = append(args, req.ProductId, req.PhoneNumber)

	default:
		return nil, fmt.Errorf("business_id yoki phone_number kerak")
	}

	row := r.Pool.QueryRow(ctx, query, args...)

	var (
		product        entity.Product
		name           sql.NullString
		categoryID     sql.NullString
		shortInfo      sql.NullString
		description    sql.NullString
		cost           sql.NullInt64
		count          sql.NullInt64
		discountCost   sql.NullInt64
		discount       sql.NullInt64
		pictureStrings sql.NullString
		productId      sql.NullInt64
	)

	err := row.Scan(
		&product.ID,
		&productId,
		&name,
		&categoryID,
		&shortInfo,
		&description,
		&cost,
		&count,
		&discountCost,
		&discount,
		&product.CreatedAt,
		&product.UpdatedAt,
		&pictureStrings,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("mahsulot topilmadi")
		}
		return nil, fmt.Errorf("ProductRepo - GetProductById - row.Scan: %w", err)
	}

	// SQL null tiplarini Product struct ga to‘g‘ri o‘tkazish
	if name.Valid {
		product.Name = name.String
	}
	if categoryID.Valid {
		product.CategoryID = categoryID.String
	}
	if shortInfo.Valid {
		product.ShortInfo = shortInfo.String
	}
	if description.Valid {
		product.Description = description.String
	}
	if cost.Valid {
		product.Cost = int(cost.Int64)
	}
	if count.Valid {
		product.Count = int(count.Int64)
	}
	if discountCost.Valid {
		product.DiscountCost = int(discountCost.Int64)
	}
	if discount.Valid {
		product.Discount = int(discount.Int64)
	}
	if pictureStrings.Valid {
		product.Image_urls = strings.Split(pictureStrings.String, ",")
	}
	if productId.Valid {
		product.ProductId = productId.Int64
	}

	// Image URL larni ajratib saqlaymiz
	if len(picturesting) > 0 {
		product.Image_urls = strings.Split(picturesting, ",")
	}

	return &product, nil
}

func (r *ProductRepo) GetProductInfoForNotification(ctx context.Context, productID string) (*entity.ProductNotificationInfo, error) {
	query := `
		SELECT name, short_info, description, cost, discount
		FROM product
		WHERE guid = $1
	`

	var (
		name        string
		shortInfo   string
		description string
		cost        int
		discount    int
	)

	err := r.Pool.QueryRow(ctx, query, productID).Scan(
		&name, &shortInfo, &description, &cost, &discount,
	)
	if err != nil {
		return nil, fmt.Errorf("ProductRepo - GetProductInfoForNotification - Scan: %w", err)
	}

	// Chegirmali narxni hisoblash
	var discountCost int
	if discount > 0 {
		discountCost = cost - (cost * discount / 100)
	} else {
		discountCost = cost
	}

	info := &entity.ProductNotificationInfo{
		Name:         name,
		ShortInfo:    shortInfo,
		Description:  description,
		Cost:         cost,
		Discount:     discount,
		DiscountCost: discountCost,
	}

	return info, nil
}

// internal/infrastructure/postgres/business.go




func (r *ProductRepo) ListProductsForAI(ctx context.Context, businessID string) ([]entity.ProductAI, error) {
	query := `
		SELECT product_id, name, cost,description,count
		FROM product
		WHERE business_id = $1 and status=true and deleted_at is null
	`

	rows, err := r.Pool.Query(ctx, query, businessID)
	if err != nil {
		fmt.Println("❌ Error while listing products for AI:", err)
		return nil, fmt.Errorf("ProductRepo - ListProductsForAI - Query: %w", err)
	}
	defer rows.Close()

	products := make([]entity.ProductAI, 0, 20)

	for rows.Next() {
		var p entity.ProductAI
		err := rows.Scan(&p.ID, &p.Name, &p.Cost, &p.Description, &p.Count)
		if err != nil {
			return nil, fmt.Errorf("ProductRepo - ListProductsForAI - rows.Scan: %w", err)
		}
		products = append(products, p)
	}

	return products, nil
}

// CheckProductCount bazadagi mahsulot soni va foydalanuvchi yuborgan sonni taqqoslaydi
func (r *ProductRepo) CheckProductCount(ctx context.Context, businessID string, products []entity.ProductOrder) (*entity.ProductCheckResponse, error) {
	fmt.Println(products)
	fmt.Println(businessID)
	for i := 0; i < len(products); i++ {
		var count int
		query := `
		SELECT count
		FROM product
		WHERE business_id = $1 and product_id = $2
	`
		row := r.Pool.QueryRow(ctx, query, businessID, products[i].ProductID)
		err := row.Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("ProductRepo - CheckProductCount - Query: %w", err)
		}
		fmt.Println(products, count)
		if products[i].Count > count {
			fmt.Println(11111111)
			return &entity.ProductCheckResponse{
				ProductID: products[i].ProductID,
				Valid:     false,
				Message:   fmt.Sprintf("Mahsulot %d uchun yetarli zaxira yo‘q. Mavjud: %d, So‘ralgan: %d", products[i].ProductID, count, products[i].Count),
			}, nil
		}
	}

	return &entity.ProductCheckResponse{
		Valid:   true,
		Message: "Mahsulot soni to‘g‘ri.",
	}, nil
}
