//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/gofer"
)

func NewPairsCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "pairs [PAIR...]",
		Aliases: []string{"pair"},
		Args:    cobra.MinimumNArgs(0),
		Short:   "List all supported asset pairs",
		Long:    `List all supported asset pairs.`,
		RunE: func(_ *cobra.Command, args []string) (err error) {
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer ctxCancel()
			sup, gof, mar, err := PrepareClientServices(ctx, opts)
			if err != nil {
				return err
			}
			if err = sup.Start(); err != nil {
				return err
			}
			defer func() { err = <-sup.Wait() }()
			defer func() {
				if err != nil {
					exitCode = 1
					_ = mar.Write(os.Stderr, err)
				}
				_ = mar.Flush()
				// Set err to nil because error was already handled by marshaller.
				err = nil
			}()
			pairs, err := gofer.NewPairs(args...)
			if err != nil {
				return err
			}
			models, err := gof.Models(pairs...)
			if err != nil {
				return err
			}
			for _, p := range models {
				if mErr := mar.Write(os.Stdout, p); mErr != nil {
					_ = mar.Write(os.Stderr, mErr)
				}
			}
			return
		},
	}
}
