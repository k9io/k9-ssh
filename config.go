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

/* k9-ssh - This program is called from sshd via the "AuthorizedKeysCommand" and
   "AuthorizedKeysCommandUser".  This uses the Key9 API to return SSH keys */

import (
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

/************************/
/* Global Configuration */
/************************/

type Configuration struct {
	System struct {
		MachineGroup      string `yaml:"machine_group"`
		RunAs             string `yaml:"run_as"`
		ConnectionTimeout int    `yaml:"connection_timeout"`
		AllowInsecure     bool   `yaml:"allow_insecure"`
	}

	Authentication struct {
		APIKey      string `yaml:"api_key"`
		CompanyUUID string `yaml:"company_uuid"`
	}

	Urls struct {
		QuerySSHKeys  string `yaml:"query_ssh_keys"`
		QueryAllUsers string `yaml:"query_all_users"`
	}
}

var Config *Configuration

func LoadConfig(configFile string) {

	/* Load config file */

	file, err := os.Open(configFile)

	if err != nil {
		log.Fatalf("Cannot open '%s' YAML file.", configFile)
	}

	defer file.Close()

	/* Init new YAML decode */

	d := yaml.NewDecoder(file)

	err = d.Decode(&Config)

	if err != nil {
		log.Fatalf("Cannot parse '%s'.", configFile)
	}

	/* Sanity Checks */

	if Config.Authentication.APIKey == "" {
		log.Fatalf("'api_key' key not found in %s.\n", configFile)
	}

	if Config.Authentication.CompanyUUID == "" {
		log.Fatalf("'company_uuid' key not found in %s.\n", configFile)
	}

	if Config.Urls.QuerySSHKeys == "" {
		log.Fatalf("'query_ssh_keys' key not found in %s.\n", configFile)
	}

	if Config.Urls.QueryAllUsers == "" {
		log.Fatalf("'query_all_users' key not found in %s.\n", configFile)
	}

	if Config.System.RunAs == "" {
		log.Fatalf("'run_as' key not found in %s.\n", configFile)
	}

	if Config.System.MachineGroup == "" {
		log.Fatalf("'machine_group' key not found in %s.\n", configFile)
	}

	/* Reject plain HTTP URLs unless explicitly permitted */

	if !Config.System.AllowInsecure {
		if strings.HasPrefix(Config.Urls.QuerySSHKeys, "http://") {
			log.Fatalf("Refusing plaintext HTTP URL for query_ssh_keys. Set allow_insecure: true to override.")
		}
		if strings.HasPrefix(Config.Urls.QueryAllUsers, "http://") {
			log.Fatalf("Refusing plaintext HTTP URL for query_all_users. Set allow_insecure: true to override.")
		}
	}

	/* Set a default value */

	if Config.System.ConnectionTimeout == 0 {
		Config.System.ConnectionTimeout = 5
	}

}
