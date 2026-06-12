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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

var validUsername = regexp.MustCompile(`^[a-z_][a-z0-9_\-]{0,31}$`)

type sshData struct {
	PublicKey string `json:"public_key"`
}

type apiError struct {
	Error string `json:"error"`
}

type allUsers struct {
	OsUsername string `json:"os_username"`
}

/********************************************************************************/
/* QueryAPI - Does the bulk of the work.  When called,  this routine makes the  */
/* API call to retreieve public keys                                             */
/********************************************************************************/

func QueryAPI(name string, remote string, displaySSHKeys bool) {

	if !validUsername.MatchString(name) {
		Log("Invalid username rejected: " + name)
		return
	}

	var postData []byte

	fullLookup := Config.Urls.QuerySSHKeys + name + "/" + Config.System.MachineGroup

	/* If we have "remote" data, send it to the API */

	if remote != "" {

		type remotePayload struct {
			Remote string `json:"remote"`
		}
		postData, _ = json.Marshal(remotePayload{Remote: remote})
		Log("Sending user '" + name + "' remote string: " + remote)

	} else {

		if displaySSHKeys {
			Log("No 'remote' string for user '" + name + ".")
		}

	}

	/* Make POST request */

	client := http.Client{Timeout: time.Duration(Config.System.ConnectionTimeout) * time.Second}

	req, err := http.NewRequest("POST", fullLookup, bytes.NewBuffer(postData))

	if err != nil {

		Log("Unable to establish API connection. Using cache for " + name)
		return

	}

	/* Send client UUID:API key */

	apiKey := fmt.Sprintf("%s:%s", Config.Authentication.CompanyUUID, Config.Authentication.APIKey)
	req.Header["API_KEY"] = []string{apiKey}

	res, err := client.Do(req)

	if err != nil {

		Log("Unable to parse keys from client.Do()")
		return

	}

	if res.StatusCode != http.StatusOK {
		Log(fmt.Sprintf("API returned unexpected status %d for user %s", res.StatusCode, name))
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {

		Log("Unable to read body from API request. Using cache for " + name)
		return
	}

	sb := string(body)

	temp := strings.Split(sb, "\n")

	data := sshData{}
	apiErr := apiError{}

	Log("Got good API response for " + name + ". Parsing keys")

	/* Public keys are dumped in JSON.  Extract public keys */

	for _, s := range temp {

		if s != "" {

			/* Check if the API returned an error object */

			if err := json.Unmarshal([]byte(s), &apiErr); err == nil && apiErr.Error != "" {
				Log("API returned error for " + name + ": " + apiErr.Error)
				return
			}

			/* If an error is hit during decode, pull the "cached" keys */

			if err := json.Unmarshal([]byte(s), &data); err != nil {

				Log("Error parsing JSON from API.  Reading cache for " + name)
				return

			}

			/* If there is no error, dump public keys from the API */

			if displaySSHKeys {
				if data.PublicKey == "" {
					Log("Warning: No key found for user " + name)
					continue
				}
				if _, _, _, _, err := gossh.ParseAuthorizedKey([]byte(data.PublicKey)); err != nil {
					Log("Rejecting malformed public key for user " + name)
					continue
				}
				fmt.Println(data.PublicKey)
			}

		}

	}

}
