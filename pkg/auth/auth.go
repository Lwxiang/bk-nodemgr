// Package auth defines the auth things in nodeman.
package auth

import (
	"git.woa.com/bk-gse/bk-nodeman/pkg/auth/authapi"
	"git.woa.com/bk-gse/bk-nodeman/pkg/auth/ssm"
)

type UserAuth struct {
	Username string

	SSM *ssm.UserAuth

	AuthAPI *authapi.UserAuth
}
