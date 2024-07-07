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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type ssh_data struct {
	PublicKey string `json:"public_key"`
}

type ssh_bypass_users struct {
	Bypass_Users string `json:"bypass_users"`
}

type all_users struct {
	Os_Username string `json:"os_username"`
}

/********************************************************************************/
/* Query_API - Does the bulk of the work.  When called,  this routine makes the */
/* API call to retreieve public keys                                            */
/********************************************************************************/

func Query_API(name string, remote string, display_ssh_keys bool) {

	var sshkeys string
	var post_data string

	full_lookup := Config.Urls.Query_SSH_Keys + name + "/" + Config.System.Machine_Group

	/* If we have "remote" data, send it to the API */

	if remote != "" {

		post_data = fmt.Sprintf("{\"remote\":\"%s\"}", remote)
		Log("Sending user '" + name + "' remote string: " + remote)

	} else {

		if display_ssh_keys == true {
			Log("No 'remote' string for user '" + name + ".")
		}

	}

	/* Make POST request */

	client := http.Client{}

	req, err := http.NewRequest("POST", full_lookup, bytes.NewBuffer([]byte(post_data)))

	if err != nil {

		Log("Unable to establish API connection. Using cache for " + name)
		sshkeys = Cache_Read(name)

		/* display_ssh_keys - dumps keys to stdout.  -precache doesnt need the
		   keys to stdout, so this can bypass that. */

		if display_ssh_keys == true {
			fmt.Print(sshkeys)
		}

		return
	}

	/* Send client UUID:API key */

	api_key_temp := fmt.Sprintf("%s:%s", Config.Authentication.Company_UUID, Config.Authentication.Api_Key)
	req.Header["API_KEY"] = []string{api_key_temp}

	res, err := client.Do(req)

	if err != nil {

		fmt.Printf("Unable to parse keys: %s\n", err)
		sshkeys = Cache_Read(name)

		if display_ssh_keys == true {
			fmt.Print(sshkeys)
		}

		return
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {

		Log("Unable to read body from API request. Using cache for " + name)
		sshkeys = Cache_Read(name)

		if display_ssh_keys == true {
			fmt.Print(sshkeys)
		}

		return
	}

	sb := string(body)

	temp := strings.Split(sb, "\n")

	sshdata := ssh_data{}

	// DEBUG: NO GOOD ERROR CHECKING! DOESNT SEE "error" KEYS.

	Log("Got good API response for " + name + ". Parsing keys")

	/* Public keys are dumped in JSON.  Extract public keys */

	for _, s := range temp {

		if s != "" {

			/* If an error is hit during decoded,  pull the "cached" keys */

			if err := json.Unmarshal([]byte(s), &sshdata); err != nil {

				Log("Error parsing JSON from API.  Reading cache for " + name)
				sshkeys = Cache_Read(name)

				if display_ssh_keys == true {
					fmt.Print(sshkeys)
				}

				return
			}

			/* If there is no error, dump public keys from the API */

			if display_ssh_keys == true {
				fmt.Println(sshdata.PublicKey)
			}

			/* Append all the keys to "sshkeys" */

			if sshdata.PublicKey != "" {
				sshkeys = sshdata.PublicKey + "\n" + sshkeys
			}

		}

	}

	/* If "sshkeys" is not blank,  write the cache out.  "Cache_Write" compares
	   previous keys to current API supplied keys to determine if a write is
	   needed */

	if sshkeys != "" {
		Cache_Write(name, sshkeys)
	}

}

/******************************************************************************/
/* PreCache - This option is run as a cron ("-precache") command line options */
/* The routine parses all users via getpwent_r() and makes API calls to       */
/* download the latest public keys.  The idea is that if the API is not       */
/* avaliable,  the user will skill be able to login via a cache public key    */
/******************************************************************************/

func PreCache() {

	allusers := all_users{}

	Log("Starting pre-cache of SSH public keys")

	client := http.Client{}

	req, err := http.NewRequest("GET", Config.Urls.Query_All_Users, nil)

	if err != nil {
		Log("Unable to establish API connection for PreCache.")
		return
	}

	api_key_temp := fmt.Sprintf("%s:%s", Config.Authentication.Company_UUID, Config.Authentication.Api_Key)
	req.Header["API_KEY"] = []string{api_key_temp}

	res, err := client.Do(req)

	if err != nil {
		Log("Unable to parse data returned from API for PreCache.")
		return
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		Log("Unable to read body for PreCache")
		return
	}

	sb := string(body)

	temp := strings.Split(sb, "\n")

	for _, s := range temp {

		if s != "" {

			if err := json.Unmarshal([]byte(s), &allusers); err != nil {

				Log("Error parsing JSON from API.")
				return

			}

			Log("Pre-caching '" + allusers.Os_Username + "'")

			Query_API(allusers.Os_Username, "", false)

		}

	}

	Log("Done pre-cache of SSH public keys")
}
