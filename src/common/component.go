package common

import "context"

type Component interface {
	Run(context.Context) error
}
