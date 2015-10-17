// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"

	idmpat "github.com/clawio/clawiod/pkg/auth/pat"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	strgpat "github.com/clawio/clawiod/pkg/storage/pat"
)

// If CreateUserHomeOnLogin is enabled it triggers the creation
// of the user home directory in
// the enabled storages.
func (a *auth) token(ctx context.Context, w http.ResponseWriter,
	r *http.Request) {

	log := logger.MustFromContext(ctx)
	idmPat := idmpat.MustFromContext(ctx)
	strgPat := strgpat.MustFromContext(ctx)

	idt, err := idmPat.Authenticate(ctx, r, r.URL.Query().Get("authtypeid"))
	if err != nil {
		log.Warning(err.Error())
		// TODO: use ValidationError/ClientError to catch 400
		// the same way with code, field and reason.
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)

		return
	}

	// Check if we have to create the user homedir in the storages.

	storages := strgPat.GetAllStorages(nil)
	for _, s := range storages {

		capParams := &storage.CapabilitiesParams{}
		capParams.Idt = idt

		if s.Capabilities(ctx, capParams).CreateUserHomeDir {

			createParams := &storage.CreateUserHomeDirParams{}
			createParams.BaseParams = capParams.BaseParams

			err := strgPat.CreateUserHomeDir(ctx, createParams, s.Prefix())
			if err != nil {
				msg := "token: creation of user home failed because err:%s"
				log.Err(fmt.Sprintf(msg, err.Error()))
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)

				return
			}
		}
	}

	tokenString, err := idmPat.CreateToken(ctx, idt)
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}

	data := make(map[string]string)
	data["authtoken"] = tokenString
	tokenJSON, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Err(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(tokenJSON)
	if err != nil {
		log.Err("token: error sending reponse. err:" + err.Error())
	}
}
