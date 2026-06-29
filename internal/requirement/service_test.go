package requirement

import (
	"context"
	"errors"
	"testing"

	"pulseroad/internal/product"
)

type stubRepo struct {
	items  map[uint]*Requirement
	nextID uint
}

func newStubRepo() *stubRepo {
	return &stubRepo{items: make(map[uint]*Requirement), nextID: 1}
}

func (r *stubRepo) Create(_ context.Context, req *Requirement) error {
	req.ID = r.nextID
	r.nextID++
	r.items[req.ID] = req
	return nil
}

func (r *stubRepo) ListByProduct(_ context.Context, productID uint, status string, page int, pageSize int) ([]Requirement, int64, error) {
	var result []Requirement
	for _, item := range r.items {
		if item.ProductID == productID {
			if status != "" && item.Status != status {
				continue
			}
			result = append(result, *item)
		}
	}
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize)
	total := int64(len(result))
	start := (page - 1) * pageSize
	if start >= len(result) {
		return nil, total, nil
	}
	end := start + pageSize
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

func (r *stubRepo) FindByID(_ context.Context, id uint) (*Requirement, error) {
	req, ok := r.items[id]
	if !ok {
		return nil, ErrRequirementNotFound
	}
	return req, nil
}

func (r *stubRepo) Update(_ context.Context, req *Requirement) error {
	r.items[req.ID] = req
	return nil
}

func (r *stubRepo) Delete(_ context.Context, id uint) error {
	if _, ok := r.items[id]; !ok {
		return ErrRequirementNotFound
	}
	delete(r.items, id)
	return nil
}

type stubProductAccess struct {
	products map[uint]uint // productID -> teamID (0 = forbidden)
}

func (a *stubProductAccess) GetProduct(_ context.Context, userID uint, productID uint) (*product.ProductResponse, error) {
	teamID, ok := a.products[productID]
	if !ok {
		return nil, product.ErrProductNotFound
	}
	if teamID == 0 {
		return nil, product.ErrForbidden
	}
	return &product.ProductResponse{ID: productID, TeamID: teamID}, nil
}

func TestCreateRequirement(t *testing.T) {
	svc := NewService(newStubRepo(), &stubProductAccess{products: map[uint]uint{1: 1, 2: 0}})
	t.Run("valid input", func(t *testing.T) {
		resp, err := svc.Create(context.Background(), 1, 1, CreateRequirementInput{
			Title:    "路线图视图",
			Priority: "p1",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Title != "路线图视图" {
			t.Errorf("title = %q, want %q", resp.Title, "路线图视图")
		}
		if resp.Status != StatusPlanned {
			t.Errorf("status = %q, want %q", resp.Status, StatusPlanned)
		}
		if resp.Priority != "p1" {
			t.Errorf("priority = %q, want %q", resp.Priority, "p1")
		}
	})
	t.Run("empty title", func(t *testing.T) {
		_, err := svc.Create(context.Background(), 1, 1, CreateRequirementInput{Title: "  "})
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("error = %v, want ErrInvalid", err)
		}
	})
	t.Run("invalid priority", func(t *testing.T) {
		_, err := svc.Create(context.Background(), 1, 1, CreateRequirementInput{Title: "x", Priority: "p99"})
		if !errors.Is(err, ErrInvalid) {
			t.Errorf("error = %v, want ErrInvalid", err)
		}
	})
	t.Run("forbidden product", func(t *testing.T) {
		_, err := svc.Create(context.Background(), 1, 2, CreateRequirementInput{Title: "x"})
		if !errors.Is(err, ErrForbidden) {
			t.Errorf("error = %v, want ErrForbidden", err)
		}
	})
}

func TestDeleteRequirement(t *testing.T) {
	repo := newStubRepo()
	svc := NewService(repo, &stubProductAccess{products: map[uint]uint{1: 1}})
	resp, _ := svc.Create(context.Background(), 1, 1, CreateRequirementInput{Title: "测试需求"})
	t.Run("creator can delete", func(t *testing.T) {
		err := svc.Delete(context.Background(), 1, resp.ID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		_, err = svc.Get(context.Background(), 1, resp.ID)
		if !errors.Is(err, ErrRequirementNotFound) {
			t.Errorf("error = %v, want ErrRequirementNotFound", err)
		}
	})
	t.Run("non-creator cannot delete", func(t *testing.T) {
		req, _ := svc.Create(context.Background(), 1, 1, CreateRequirementInput{Title: "另一条"})
		err := svc.Delete(context.Background(), 2, req.ID)
		if !errors.Is(err, ErrNotOwner) {
			t.Errorf("error = %v, want ErrNotOwner", err)
		}
	})
}
