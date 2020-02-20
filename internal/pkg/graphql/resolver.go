// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.
package graphql

import (
	"context"
)

type Resolver struct{}

func (r *entityResolver) FindDeviceByID(ctx context.Context, id string) (*Device, error) {
	return &Device{ID: id}, nil
}

func (r *queryResolver) Temperatures(ctx context.Context) ([]*Temperature, error) {
	return []*Temperature{}, nil
}

func (r *Resolver) Entity() EntityResolver { return &entityResolver{r} }
func (r *Resolver) Query() QueryResolver   { return &queryResolver{r} }

type entityResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
