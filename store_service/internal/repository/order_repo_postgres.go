package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OrderRepoPostgres struct {
	db  PgxIface
	log *logger.Logger
}

func NewOrderRepoPostgres(db PgxIface, log *logger.Logger) *OrderRepoPostgres {
	return &OrderRepoPostgres{
		db:  db,
		log: log,
	}
}

func (r *OrderRepoPostgres) CreateOrder(ctx context.Context, userID string) (string, error) {
	r.log.Debug(ctx, "CreateOrder начало обработки", map[string]interface{}{})

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// создаем пустую запись
	query := `
		insert into "order" (id, user_id, total_price)
		values ($1, $2, 0);
	`
	orderID := uuid.New().String()
	_, err = tx.Exec(ctx, query, orderID, userID)
	if err != nil {
		r.log.Error(ctx, "CreateOrder ошибка создания заказа", map[string]interface{}{"err": err, "id": userID})
		return "", err
	}

	// переносим корзину
	query = `
		insert into order_item (id, order_id, store_item_id, price, quantity)
		select gen_random_uuid(), $1, si.id, si.price, ci.quantity
		from cart_item ci
		join cart c on c.id = ci.cart_id
		join store_item si on si.id = ci.store_item_id
		where c.user_id = $2;
	`
	_, err = tx.Exec(ctx, query, orderID, userID)
	if err != nil {
		r.log.Error(ctx, "CreateOrder ошибка переноса товаров", map[string]interface{}{"err": err, "id": userID})
		return "", err
	}

	// сумма заказа
	query = `
		update "order"
		set total_price = (
			select sum(si.price * ci.quantity)
			from cart_item ci
			join cart c ON c.id = ci.cart_id
			join store_item si ON si.id = ci.store_item_id
			where c.user_id = $1
		)
		where id = $2;
	`
	_, err = tx.Exec(ctx, query, userID)
	if err != nil {
		r.log.Error(ctx, "CreateOrder ошибка обновления суммы", map[string]interface{}{"err": err, "id": userID})
		return "", err
	}

	// очистка корзины
	query = `
		delete from cart_item
		where cart_id = (select id from cart where user_id = $1);
	`
	_, err = tx.Exec(ctx, query, userID)
	if err != nil {
		r.log.Error(ctx, "CreateOrder ошибка удаления записи", map[string]interface{}{"err": err, "id": userID})
		return "", err
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.log.Error(ctx, "CreateOrder ошибка закрытия транзакции", map[string]interface{}{"err": err, "id": userID})
		return "", domain.ErrInternalServer
	}

	r.log.Debug(ctx, "CreateOrder завершено успешно", map[string]interface{}{})
	return orderID, nil
}

func (r *OrderRepoPostgres) UpdateOrderStatus(ctx context.Context, id, status string) error {
	r.log.Debug(ctx, "UpdateOrderStatus начало обработки", map[string]interface{}{})

	query := `
		update "order"
		set status = $2
		where id = $1
		returning id
	`

	var orderID string
	err := r.db.QueryRow(ctx, query, id, status).Scan(&orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warn(ctx, "UpdateOrderStatus заказ отсутствует",
				map[string]interface{}{"err": err, "id": id, "status": status})
			return domain.ErrRowsNotFound
		}
		r.log.Error(ctx, "UpdateOrderStatus ошибка обновления статуса",
			map[string]interface{}{"err": err, "id": id, "status": status})
		return err
	}

	r.log.Debug(ctx, "UpdateOrderStatus завершено успешно", map[string]interface{}{})
	return nil
}

func (r *OrderRepoPostgres) GetOrder(ctx context.Context, id string) (*domain.OrderInfo, error) {
	r.log.Debug(ctx, "GetOrder начало обработки", map[string]interface{}{})

	query := `
		select
			o.id as order_id,
			o.total_price as total,
			o.status as status,
			o.created_at as created_at,
			si.id as store_item_id,
			i.name as name,
			i.card_img as card_img,
			oi.price as price,
			oi.quantity as quantity
		from "order" o
		join order_item oi on oi.order_id = o.id
		join store_item si on si.id = oi.store_item_id
		join item i on i.id = si.item_id
		where o.id = $1
		order by oi.created_at;
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		r.log.Error(ctx, "GetOrder ошибка бд", map[string]interface{}{"err": err, "id": id})
		return nil, err
	}
	defer rows.Close()

	var order domain.OrderInfo
	var items []*domain.OrderItemInfo
	for rows.Next() {
		var item domain.OrderItemInfo
		err = rows.Scan(
			&order.ID,
			&order.Total,
			&order.Status,
			&order.CreatedAt,
			&item.ID,
			&item.Name,
			&item.CardImg,
			&item.Price,
			&item.Quantity,
		)
		if err != nil {
			r.log.Error(ctx, "GetOrder ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		items = append(items, &item)
	}

	err = rows.Err()
	if err != nil {
		r.log.Error(ctx, "GetOrder ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": id})
		return nil, err
	}

	if len(items) == 0 {
		r.log.Warn(ctx, "GetOrder пустой ответ", map[string]interface{}{"id": id})
		return nil, domain.ErrRowsNotFound
	}

	order.Items = items
	r.log.Debug(ctx, "GetOrder завершено успешно", map[string]interface{}{})
	return &order, nil
}

func (r *OrderRepoPostgres) GetOrdersUser(ctx context.Context, userID string) ([]*domain.Order, error) {
	r.log.Debug(ctx, "GetOrdersUser начало обработки", map[string]interface{}{})

	query := `
		select
			o.id as id,
			o.status as status,
			o.total_price as total,
			o.created_at as created_at
		from "order" o
		where o.user_id = $1
		order by o.created_at;
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.log.Error(ctx, "GetOrdersUser ошибка бд", map[string]interface{}{"err": err, "id": userID})
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var order domain.Order
		err = rows.Scan(
			&order.ID,
			&order.Status,
			&order.Total,
			&order.CreatedAt,
		)
		if err != nil {
			r.log.Error(ctx, "GetOrdersUser ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		orders = append(orders, &order)
	}

	err = rows.Err()
	if err != nil {
		r.log.Error(ctx, "GetOrdersUser ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": userID})
		return nil, err
	}

	if len(orders) == 0 {
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetOrdersUser завершено успешно", map[string]interface{}{})
	return orders, nil
}
