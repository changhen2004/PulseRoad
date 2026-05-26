package product

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeProductRepository struct {
	nextID   uint
	products map[uint]*Product
}

func newFakeProductRepository() *fakeProductRepository {
	return &fakeProductRepository{
		nextID:   1,
		products: make(map[uint]*Product),
	}
}

func (r *fakeProductRepository) Create(_ context.Context, product *Product) error {
	product.ID = r.nextID
	product.CreatedAt = time.Now()
	product.UpdatedAt = product.CreatedAt
	r.nextID++

	copy := *product
	r.products[product.ID] = &copy
	return nil
}

func (r *fakeProductRepository) ListByTeam(_ context.Context, teamID uint) ([]Product, error) {
	var products []Product
	for _, product := range r.products {
		if product.TeamID == teamID {
			products = append(products, *product)
		}
	}
	return products, nil
}

func (r *fakeProductRepository) FindByID(_ context.Context, id uint) (*Product, error) {
	product, ok := r.products[id]
	if !ok {
		return nil, ErrProductNotFound
	}
	copy := *product
	return &copy, nil
}

type fakeTeamMembership struct {
	members map[uint]map[uint]bool
}

func newFakeTeamMembership() *fakeTeamMembership {
	return &fakeTeamMembership{members: make(map[uint]map[uint]bool)}
}

func (m *fakeTeamMembership) add(teamID uint, userID uint) {
	if m.members[teamID] == nil {
		m.members[teamID] = make(map[uint]bool)
	}
	m.members[teamID][userID] = true
}

func (m *fakeTeamMembership) IsMember(_ context.Context, userID uint, teamID uint) (bool, error) {
	return m.members[teamID][userID], nil
}

func TestCreateProductRequiresTeamMember(t *testing.T) {
	membership := newFakeTeamMembership()
	membership.add(10, 7)
	svc := NewService(newFakeProductRepository(), membership)

	created, err := svc.CreateProduct(context.Background(), 7, 10, CreateProductInput{
		Name:        "PulseRoad",
		Description: "Feedback platform",
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	if created.ID == 0 {
		t.Fatal("expected product id")
	}
	if created.TeamID != 10 {
		t.Fatalf("expected team id 10, got %d", created.TeamID)
	}
	if created.Name != "PulseRoad" || created.CreatedBy != 7 {
		t.Fatalf("unexpected product: %#v", created)
	}
}

func TestCreateProductRejectsNonMember(t *testing.T) {
	svc := NewService(newFakeProductRepository(), newFakeTeamMembership())

	_, err := svc.CreateProduct(context.Background(), 7, 10, CreateProductInput{Name: "PulseRoad"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestListProductsRequiresTeamMember(t *testing.T) {
	membership := newFakeTeamMembership()
	membership.add(10, 7)
	repo := newFakeProductRepository()
	svc := NewService(repo, membership)

	if _, err := svc.CreateProduct(context.Background(), 7, 10, CreateProductInput{Name: "PulseRoad"}); err != nil {
		t.Fatalf("create product: %v", err)
	}
	if _, err := svc.CreateProduct(context.Background(), 7, 10, CreateProductInput{Name: "FlagFlow"}); err != nil {
		t.Fatalf("create second product: %v", err)
	}

	products, err := svc.ListProducts(context.Background(), 7, 10)
	if err != nil {
		t.Fatalf("list products: %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected two products, got %#v", products)
	}

	_, err = svc.ListProducts(context.Background(), 8, 10)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden for non-member, got %v", err)
	}
}

func TestGetProductRejectsNonMember(t *testing.T) {
	membership := newFakeTeamMembership()
	membership.add(10, 7)
	repo := newFakeProductRepository()
	svc := NewService(repo, membership)

	created, err := svc.CreateProduct(context.Background(), 7, 10, CreateProductInput{Name: "PulseRoad"})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	_, err = svc.GetProduct(context.Background(), 8, created.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}

	product, err := svc.GetProduct(context.Background(), 7, created.ID)
	if err != nil {
		t.Fatalf("get product as member: %v", err)
	}
	if product.ID != created.ID || product.TeamID != 10 {
		t.Fatalf("unexpected product detail: %#v", product)
	}
}
