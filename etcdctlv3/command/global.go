// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/spf13/cobra"
)

// GlobalFlags are flags that defined globally
// and are inherited to all sub-commands.
type GlobalFlags struct {
	Endpoints   []string
	DialTimeout time.Duration

	TLS transport.TLSInfo

	OutputFormat string
	IsHex        bool
}

var display printer = &simplePrinter{}

func mustClientFromCmd(cmd *cobra.Command) *clientv3.Client {
	endpoints, err := cmd.Flags().GetStringSlice("endpoints")
	if err != nil {
		ExitWithError(ExitError, err)
	}
	dialTimeout := dialTimeoutFromCmd(cmd)
	cert, key, cacert := keyAndCertFromCmd(cmd)

	return mustClient(endpoints, dialTimeout, cert, key, cacert)
}

func mustClient(endpoints []string, dialTimeout time.Duration, cert, key, cacert string) *clientv3.Client {
	cfg, err := newClientCfg(endpoints, dialTimeout, cert, key, cacert)
	if err != nil {
		ExitWithError(ExitBadArgs, err)
	}

	client, err := clientv3.New(*cfg)
	if err != nil {
		ExitWithError(ExitBadConnection, err)
	}

	return client
}

func newClientCfg(endpoints []string, dialTimeout time.Duration, cert, key, cacert string) (*clientv3.Config, error) {
	// set tls if any one tls option set
	var cfgtls *transport.TLSInfo
	tls := transport.TLSInfo{}
	var file string
	if cert != "" {
		tls.CertFile = cert
		cfgtls = &tls
	}

	if key != "" {
		tls.KeyFile = key
		cfgtls = &tls
	}

	if cacert != "" {
		tls.CAFile = file
		cfgtls = &tls
	}

	cfg := &clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}
	if cfgtls != nil {
		clientTLS, err := cfgtls.ClientConfig()
		if err != nil {
			return nil, err
		}
		cfg.TLS = clientTLS
	}

	return cfg, nil
}

func argOrStdin(args []string, stdin io.Reader, i int) (string, error) {
	if i < len(args) {
		return args[i], nil
	}
	bytes, err := ioutil.ReadAll(stdin)
	if string(bytes) == "" || err != nil {
		return "", errors.New("no available argument and stdin")
	}
	return string(bytes), nil
}

func dialTimeoutFromCmd(cmd *cobra.Command) time.Duration {
	dialTimeout, err := cmd.Flags().GetDuration("dial-timeout")
	if err != nil {
		ExitWithError(ExitError, err)
	}
	return dialTimeout
}

func keyAndCertFromCmd(cmd *cobra.Command) (cert, key, cacert string) {
	var err error
	if cert, err = cmd.Flags().GetString("cert"); err != nil {
		ExitWithError(ExitBadArgs, err)
	} else if cert == "" && cmd.Flags().Changed("cert") {
		ExitWithError(ExitBadArgs, errors.New("empty string is passed to --cert option"))
	}

	if key, err = cmd.Flags().GetString("key"); err != nil {
		ExitWithError(ExitBadArgs, err)
	} else if key == "" && cmd.Flags().Changed("key") {
		ExitWithError(ExitBadArgs, errors.New("empty string is passed to --key option"))
	}

	if cacert, err = cmd.Flags().GetString("cacert"); err != nil {
		ExitWithError(ExitBadArgs, err)
	} else if cacert == "" && cmd.Flags().Changed("cacert") {
		ExitWithError(ExitBadArgs, errors.New("empty string is passed to --cacert option"))
	}

	return cert, key, cacert
}
