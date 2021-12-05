package ident

import (
	"github.com/rs/xid"
)

func UniqueID() string {
	return xid.New().String()
}
