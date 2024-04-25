package connector

import "github.com/YAOHAO9/pine/rpc/session"

type Options struct {
	authCb  func(uid, token string, sessionData map[string]string) error
	closeCb func(session *session.Session, err error)
}

type Option func(o *Options)

func WithOnAuthConnFn(authFn func(uid, token string, initSessionData map[string]string) error) Option {
	return func(o *Options) {
		o.authCb = authFn
	}
}

func WithOnCloseConnFn(closeFn func(session *session.Session, err error)) Option {
	return func(o *Options) {
		o.closeCb = closeFn
	}
}