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

/* Write/reads a "cached" copy of the SSH keys if we can't connect to the
   Key9 API.  If this is something your worried about, you should consider
   running the k9-proxy (https://github.com/k9io/k9-proxy) */

import (
	"bytes"
	"crypto/sha256"
	"os"
)

func Cache_Read(name string) string {

	Cache_User_File := Config.System.Cache_Dir + "/" + name

	_, err := os.Stat(Cache_User_File)

	if err != nil {
		Log("Can't read cache file: " + Cache_User_File)
		os.Exit(-1)
	}

	cached_sshkeys, err := os.ReadFile(Cache_User_File)

	if err != nil {
		Log("Can't read cache file: " + Cache_User_File)
		os.Exit(-1)
	}

	return (string(cached_sshkeys))

}

func Cache_Write(name string, sshkeys string) {

	Cache_User_File := Config.System.Cache_Dir + "/" + name

	/* If user cache doesn't exist, then create it! */

	_, err_stat := os.Stat(Cache_User_File)

	if err_stat != nil {

		Log("No cache file detected, creating " + Cache_User_File)

		err_write := os.WriteFile(Cache_User_File, []byte(sshkeys), 0600)

		if err_write != nil {

			Log("Can't write cache file " + Cache_User_File)
			os.Exit(-1)

		}

		return /* Don't need to go any further */

	}

	/* Current passkey hash */

	current_sshkeys_hash := sha256.Sum256([]byte(sshkeys))

	/* Grab the cached sshkeys */

	cached_sshkeys := Cache_Read(name)
	cached_sshkeys_hash := sha256.Sum256([]byte(cached_sshkeys))

	/* Passkey have changed,  let's make a copy for cache */

	if bytes.Equal(current_sshkeys_hash[:], cached_sshkeys_hash[:]) == false {

		Log("Keys have been updated, writing new cache file " + Cache_User_File)

		err_write := os.WriteFile(Cache_User_File, []byte(sshkeys), 0600)

		if err_write != nil {
			Log("Can't write cache file " + Cache_User_File)
			os.Exit(-1)

		}

	}

}
