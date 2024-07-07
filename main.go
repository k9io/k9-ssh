/*
 * Copyright (C) 2024 Key9 Identity, Inc <k9.io>
 * Copyright (C) 2024 Champ Clark III <cclark@k9.io>
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

func main() {

	username := flag.String("user", "", "System username to query.")
	precache := flag.Bool("precache", false, "Pre-cache all users.")
	remote := flag.String("remote", "", "Remote data.")

	flag.Parse()

	/* Sanity check */

	if *username == "" && *precache == false {
		Log("No username specified!")
		os.Exit(1)
	}

	/* Load configuration */

	LoadConfig("/opt/k9/etc/k9.yaml")

	/* See if the proper user is executing this routine.  We want to keep this locked down
	   as much as possible */

	C, err := user.Current()

	if err != nil {
		Log("Unable to determine username executing k9-ssh")
		os.Exit(-1)
	}

	if C.Username != Config.System.Run_As {
		Log("User " + C.Username + "does not matched expected user " + Config.System.Run_As)
		os.Exit(-1)
	}

	/* Do API calls */

	if *username != "" {
		Query_API(*username, *remote, true)
	}

	if *precache == true {
		PreCache()
	}

}
