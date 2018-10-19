// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package firebase

import (
	"context"
	"fmt"
	"io"
	"os"

	"firebase.google.com/go"
	"google.golang.org/api/option"

	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

//Start starts the file module
func Start(in io.Reader, out io.Writer) {
	module := abm.Register("firebase")

	module.FileLoaderRegister("path", firebasePath)
	module.FileParserRegister("get", getFromFirebase)

	module.CustomStart(in, out)
}

func firebasePath(w abm.FLRequest, r *abm.FLResponse) {
	if w.Variable == "" {
		r.Error = abm.ErrInvalidInput
		return
	}

	r.Data = []byte(w.Variable)
}

func getFromFirebase(w abm.FPRequest, r *abm.FPResponse) {
	if w.Data == nil {
		r.Error = abm.ErrInvalidInput
		fmt.Fprintln(os.Stderr, r.Error)
		return
	}

	if w.Variable == "" {
		r.Error = abm.ErrInvalidInput
		fmt.Fprintln(os.Stderr, r.Error)
		return
	}

	fmt.Fprintln(os.Stderr, w.Data, w.Variable)

	opt := option.WithCredentialsFile(w.Variable)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		r.Error = err
		fmt.Fprintln(os.Stderr, r.Error)
		return
	}

	firestore, err := app.Firestore(context.Background())
	if err != nil {
		r.Error = err
		fmt.Fprintln(os.Stderr, r.Error)
		return
	}

	frl := w.Data

	data, err := firestore.Doc(string(frl)).Get(context.Background())
	if err != nil {
		r.Error = err
		fmt.Fprintln(os.Stderr, r.Error)
		return
	}

	if !data.Exists() {
		r.Error = abm.ErrFailed
		fmt.Fprintln(os.Stderr, r.Error)
		return
	}

	var firebaseData = data.Data()

	if err != nil {
		r.Error = err
		fmt.Fprintln(os.Stderr, err)
		return
	}

	r.Data = firebaseData
}
