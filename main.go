/*
 * Copyright (C) 2024-2025 Key9 Identity, Inc <k9.io>
 * Copyright (C) 2024-2025 Champ Clark III <cclark@k9.io>
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License Version 2 as
 * published by the Free Software Foundation.  You may not use, modify or
 * distribute this program under any other version of the GNU General
 * Public License.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
 */

package main

import (
	"flag"
	"os"
	"os/user"
)

var version = "1.1-0"

func main() {

	username := flag.String("user", "", "System username to query.")
	precache := flag.Bool("precache", false, "Pre-cache all users.")
	remote := flag.String("remote", "", "Remote data.")
	configFile := flag.String("config", "", "Configuration file.")

	flag.Parse()

	initLog()

	/* Sanity check */

	if *configFile == "" {
		*configFile = "/opt/k9/etc/k9.yaml" /* Set default */
	}

	if *username == "" && !*precache {
		Log("No username specified!")
		os.Exit(1)
	}

	/* Load configuration */

	LoadConfig(*configFile)

	/* See if the proper user is executing this routine.  We want to keep this locked down as much as possible */

	c, err := user.Current()

	if err != nil {
		Log("Unable to determine username executing k9-ssh")
		os.Exit(1)
	}

	if c.Username != Config.System.RunAs {
		Log("User " + c.Username + " does not match expected user " + Config.System.RunAs)
		os.Exit(1)
	}

	/* Do API calls */

	if *username != "" {
		QueryAPI(*username, *remote, true)
	}

	if *precache {
		PreCache()
	}

}
