package usergrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	userCore "service/domain/core/user"
	"service/domain/data/store/user"
	"service/domain/sys/auth"
	"service/domain/sys/database"
	"service/domain/sys/validate"
	"service/foundation/web"
	"strconv"
)

type Handlers struct {
	Core userCore.Core
	Auth *auth.Auth
}

func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	page := web.Param(r, "page")
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid page format [%s] ", err), http.StatusBadRequest)
	}

	rows := web.Param(r, "rows")
	rowNum, err := strconv.Atoi(rows)
	if err != nil {
		return validate.NewRequestError(fmt.Errorf("invalid rows format [%s] ", err), http.StatusBadRequest)
	}

	usr, err := h.Core.Query(ctx, pageNum, rowNum)
	if err != nil {
		return fmt.Errorf("unable to query for users [%w] ", err)
	}

	return web.Respond(ctx, w, http.StatusOK, usr)
}

func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return errors.New("claims are missing from context ")
	}

	id := web.Param(r, "id")
	usr, err := h.Core.QueryByID(ctx, claims, id)
	if err != nil {
		switch validate.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		case database.ErrForbidden:
			return validate.NewRequestError(err, http.StatusForbidden)
		default:
			return fmt.Errorf("ID[%s] %w", id, err)
		}
	}
	return web.Respond(ctx, w, http.StatusOK, usr)
}

func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	//TODO min 10
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web values missing from content")
	}

	var nu user.NewUser
	if err := web.Decode(r, nu); err != nil {
		return fmt.Errorf("unable to decode payload  %w", err)
	}

	usr, err := h.Core.Create(ctx, nu, v.Now)
	if err != nil {
		return fmt.Errorf("user %+v %w", &usr, err)
	}
	return web.Respond(ctx, w, http.StatusCreated, usr)
}

func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web values missing from content")
	}

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return errors.New("claims are missing from context ")
	}

	var upd user.UpdateUser
	if err := web.Decode(r, upd); err != nil {
		return fmt.Errorf("unable to decode payload  %w", err)
	}

	id := web.Param(r, "id")
	err = h.Core.Update(ctx, claims, id, upd, v.Now)
	if err != nil {
		switch validate.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		case database.ErrForbidden:
			return validate.NewRequestError(err, http.StatusForbidden)
		default:
			return fmt.Errorf("ID[%s] User[%+v] %w", id, &upd, err)
		}
	}
	return web.Respond(ctx, w, http.StatusOK, upd)

}

func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return errors.New("claims are missing from context ")
	}

	id := web.Param(r, "id")
	err = h.Core.Delete(ctx, claims, id)
	if err != nil {
		switch validate.Cause(err) {
		case database.ErrInvalidID:
			return validate.NewRequestError(err, http.StatusBadRequest)
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		case database.ErrForbidden:
			return validate.NewRequestError(err, http.StatusForbidden)
		default:
			return fmt.Errorf("ID[%s] User[%+v] %w", id, err)
		}
	}
	return web.Respond(ctx, w, http.StatusOK, nil)
}

func (h Handlers) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web values missing from content")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in basic auth")
		return validate.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := h.Core.Authenticate(ctx, v.Now, email, pass)
	if err != nil {
		switch validate.Cause(err) {
		case database.ErrNotFound:
			return validate.NewRequestError(err, http.StatusNotFound)
		case database.ErrAuthenticationFailure:
			return validate.NewRequestError(err, http.StatusUnauthorized)
		default:
			return fmt.Errorf("authenticating ... %w", err)
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = h.Auth.GenerateToken(claims)
	if err != nil {
		return fmt.Errorf("generating token  %w", err)
	}
	return web.Respond(ctx, w, http.StatusOK, tkn)
}
