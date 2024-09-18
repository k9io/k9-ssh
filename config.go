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

/* k9-ssh - This program is called from sshd via the "AuthorizedKeysCommand" and
   "AuthorizedKeysCommandUser".  This uses the Key9 API to return SSH keys */

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

/************************/
/* Global Configuration */
/************************/

type Configuration struct {
	System struct {
		Machine_Group string `yaml:"machine_group"`
		Run_As        string `yaml:"run_as"`
		Cache_Dir     string `yaml:"cache_dir"`
		Connection_Timeout int `yaml:"connection_timeout"`
	}

	Authentication struct {
		Api_Key      string `yaml:"api_key"`
		Company_UUID string `yaml:"company_uuid"`
	}

	Urls struct {
		Query_SSH_Keys  string `yaml:"query_ssh_keys"`
		Query_All_Users string `yaml:"query_all_users"`
	}
}

var Config *Configuration

func LoadConfig(config_file string) {

	/* Load config file */

	file, err := os.Open(config_file)

	if err != nil {
		log.Fatalf("Cannot open '%s' YAML file.", config_file)
	}

	defer file.Close()

	/* Init new YAML decode */

	d := yaml.NewDecoder(file)

	err = d.Decode(&Config)

	if err != nil {
		log.Fatalf("Cannot parse '%s'.", config_file)
	}

	/* Sanity Checks */

	if Config.Authentication.Api_Key == "" {
		log.Fatalf("'api_key' key not found in %s.\n", config_file)
	}

	if Config.Authentication.Company_UUID == "" {
		log.Fatalf("'company_uuid' key not found in %s.\n", config_file)
	}

	if Config.Urls.Query_SSH_Keys == "" {
		log.Fatalf("'query_ssh_keys' key not found in %s.\n", config_file)
	}

	if Config.Urls.Query_All_Users == "" {
		log.Fatalf("'query_k9_allusers' key not found in %s.\n", config_file)
	}

	if Config.System.Run_As == "" {
		log.Fatalf("'run_as' key not found in %s.\n", config_file)
	}

	if Config.System.Cache_Dir == "" {
		log.Fatalf("'cache_dir' key not found in %s.\n", config_file)
	}

	if Config.System.Machine_Group == "" {
		log.Fatalf("'group' key not found in %s.\n", config_file)
	}

	/* Set a default value */

	if Config.System.Connection_Timeout == 0 {
		Config.System.Connection_Timeout = 5
	}

}
