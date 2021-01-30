package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/levigross/grequests"

	_ "github.com/lib/pq"
)

func runEnvCmd(env []string, command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Executing command:", command, "with args:", arg)
	return cmd.Run()
}

func runCmd(command string, arg ...string) error {
	return runEnvCmd([]string{}, command, arg...)
}

func checkOutput(command string, arg ...string) string {
	log.Println("Executing command:", command, "with args:", arg)
	out, err := exec.Command(command, arg...).Output()
	check(err)
	return strings.TrimSpace(string(out))
}

func mkdirs(mode os.FileMode, arg ...string) string {
	path := ""
	for _, folder := range arg {
		path = filepath.Join(path, folder)
		os.MkdirAll(path, mode)
		os.Chmod(path, mode)
	}
	return path
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func initDocker(composerTemplate string, env []string) {
	if runtime.GOOS == "linux" {
		// set map_max_count size to 262144
		check(runCmd("sysctl", "-w", "vm.max_map_count=262144"))
		f, err := os.OpenFile("/etc/sysctl.conf", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		check(err)
		defer f.Close()
		f.WriteString("vm.max_map_count=262144")
	}

	if runCmd("docker", "version") != nil {
		installDocker()
	}
	runCmd("docker", "swarm", "init")

	// login to registry
	runCmd("docker", "login", "-u", "client", "-p", "4xYkVIAH8kdAH7mP/9BBhbb2ByzLGm4F", "utmstack.azurecr.io")

	// pull images from registry
	for _, image := range containersImages {
		check(runCmd("docker", "pull", "utmstack.azurecr.io/"+image))
	}

	// generate composer file and deploy
	f, err := os.Create(composerFile)
	check(err)
	defer f.Close()
	f.WriteString(composerTemplate)

	check(runEnvCmd(env, "docker", "stack", "deploy", "--compose-file", composerFile, stackName))
}

func generateCerts(nginxCert string){
	cert := `-----BEGIN CERTIFICATE-----
MIIFJjCCBA6gAwIBAgISA7ylpw0Ob1YkGwHhx3lwj3gwMA0GCSqGSIb3DQEBCwUA
MDIxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MQswCQYDVQQD
EwJSMzAeFw0yMTAxMDExNDMwMTRaFw0yMTA0MDExNDMwMTRaMBwxGjAYBgNVBAMT
EXNpZW0udXRtc3RhY2suY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC
AQEAwOCCCS37KDMRpsW36wBt80hz4G5HCHnQZhvkpPE3eSGTAHYu1H4ZoPVpe/47
qf1qxW3+Y4dzICn4KpeBSiDhC4L4vU/hEmjrhtpQgy+7Acxqk3WmlJ9i/fSsTiwA
ahyIGdyR1FDW1WzpT5nCF2v+xVQM2vW3HOCa2P4M0Edc7T8ki7ZBuonc3ViMCc8K
yuXQ37auVVU6znH1Nm/MhimSFGvf6ZKVRm/ANPtMnnrw6J2IZ4YtPhzfBJ1fWNYg
5FqI/+iZ/HE/gxYxPowb9RyEFi7y8OWoD+j+8c1PFlfq9nGcEOtw/LWsnwvG2nnS
afQ89BuSxW5xf3Er3P4KqV6/jQIDAQABo4ICSjCCAkYwDgYDVR0PAQH/BAQDAgWg
MB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMB0G
A1UdDgQWBBStPJIHEpzk/cLXMUjzToJEWcxFFzAfBgNVHSMEGDAWgBQULrMXt1hW
y65QCUDmH6+dixTCxjBVBggrBgEFBQcBAQRJMEcwIQYIKwYBBQUHMAGGFWh0dHA6
Ly9yMy5vLmxlbmNyLm9yZzAiBggrBgEFBQcwAoYWaHR0cDovL3IzLmkubGVuY3Iu
b3JnLzAcBgNVHREEFTATghFzaWVtLnV0bXN0YWNrLmNvbTBMBgNVHSAERTBDMAgG
BmeBDAECATA3BgsrBgEEAYLfEwEBATAoMCYGCCsGAQUFBwIBFhpodHRwOi8vY3Bz
LmxldHNlbmNyeXB0Lm9yZzCCAQIGCisGAQQB1nkCBAIEgfMEgfAA7gB1AG9Tdqwx
8DEZ2JkApFEV/3cVHBHZAsEAKQaNsgiaN9kTAAABdr6SGvEAAAQDAEYwRAIgL/0c
Q+x3pzOGj8V6ri9SuAq6D3isp2DH2jXaxtVzz5MCIGbiBtVeZVd2X1IBCokHepnX
YeDQsFGAQFytPIhLxllLAHUA9lyUL9F3MCIUVBgIMJRWjuNNExkzv98MLyALzE7x
ZOMAAAF2vpIa2gAABAMARjBEAiA6zpHXvk7WW+EFAcNDwsxUxERrZESge4REj+TN
6hL97wIgXYEsqAFAzvd6/zT2LujrjtYiLXYsnY6lhQQHjb4FE/kwDQYJKoZIhvcN
AQELBQADggEBACPPScgPwbItxm0c7IADO4BZAdAofcZXUNmgdTG7a/N1+c5wbEat
D3Q+2tIwaoo10NLWodncIdazL/zFxCjxlw3V8ms2/Hq4Y+eCelsgtkinll7cMiUO
1g0YwSRhfx1bZG8VLmyp2Mt8tP1t3fo//pKz8///mdfN31FQjP7n3w3EM+eWQNtc
ZkZKOE10irEjZwHs1np49iqlq0y1hw7eOdVylZSBmnKu+EGTLlpq4xKHIwNpi6B8
+Fz4r/BzbcvH64B++rFxjp8qFcHvNAUCMDCASSX5DTn0lOxr6m5wuHUmQ47kwdBZ
5GbDRHdkYFV3FoteCj+ztlIx1iVY/tJRYUQ=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEZTCCA02gAwIBAgIQQAF1BIMUpMghjISpDBbN3zANBgkqhkiG9w0BAQsFADA/
MSQwIgYDVQQKExtEaWdpdGFsIFNpZ25hdHVyZSBUcnVzdCBDby4xFzAVBgNVBAMT
DkRTVCBSb290IENBIFgzMB4XDTIwMTAwNzE5MjE0MFoXDTIxMDkyOTE5MjE0MFow
MjELMAkGA1UEBhMCVVMxFjAUBgNVBAoTDUxldCdzIEVuY3J5cHQxCzAJBgNVBAMT
AlIzMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuwIVKMz2oJTTDxLs
jVWSw/iC8ZmmekKIp10mqrUrucVMsa+Oa/l1yKPXD0eUFFU1V4yeqKI5GfWCPEKp
Tm71O8Mu243AsFzzWTjn7c9p8FoLG77AlCQlh/o3cbMT5xys4Zvv2+Q7RVJFlqnB
U840yFLuta7tj95gcOKlVKu2bQ6XpUA0ayvTvGbrZjR8+muLj1cpmfgwF126cm/7
gcWt0oZYPRfH5wm78Sv3htzB2nFd1EbjzK0lwYi8YGd1ZrPxGPeiXOZT/zqItkel
/xMY6pgJdz+dU/nPAeX1pnAXFK9jpP+Zs5Od3FOnBv5IhR2haa4ldbsTzFID9e1R
oYvbFQIDAQABo4IBaDCCAWQwEgYDVR0TAQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8E
BAMCAYYwSwYIKwYBBQUHAQEEPzA9MDsGCCsGAQUFBzAChi9odHRwOi8vYXBwcy5p
ZGVudHJ1c3QuY29tL3Jvb3RzL2RzdHJvb3RjYXgzLnA3YzAfBgNVHSMEGDAWgBTE
p7Gkeyxx+tvhS5B1/8QVYIWJEDBUBgNVHSAETTBLMAgGBmeBDAECATA/BgsrBgEE
AYLfEwEBATAwMC4GCCsGAQUFBwIBFiJodHRwOi8vY3BzLnJvb3QteDEubGV0c2Vu
Y3J5cHQub3JnMDwGA1UdHwQ1MDMwMaAvoC2GK2h0dHA6Ly9jcmwuaWRlbnRydXN0
LmNvbS9EU1RST09UQ0FYM0NSTC5jcmwwHQYDVR0OBBYEFBQusxe3WFbLrlAJQOYf
r52LFMLGMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjANBgkqhkiG9w0B
AQsFAAOCAQEA2UzgyfWEiDcx27sT4rP8i2tiEmxYt0l+PAK3qB8oYevO4C5z70kH
ejWEHx2taPDY/laBL21/WKZuNTYQHHPD5b1tXgHXbnL7KqC401dk5VvCadTQsvd8
S8MXjohyc9z9/G2948kLjmE6Flh9dDYrVYA9x2O+hEPGOaEOa1eePynBgPayvUfL
qjBstzLhWVQLGAkXXmNs+5ZnPBxzDJOLxhF2JIbeQAcH5H0tZrUlo5ZYyOqA7s9p
O5b85o3AM/OJ+CktFBQtfvBhcJVd9wvlwPsk+uyOy2HI7mNxKKgsBTt375teA2Tw
UdHkhVNcsAKX1H7GNNLOEADksd86wuoXvg==
-----END CERTIFICATE-----
`
	key := `-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDA4IIJLfsoMxGm
xbfrAG3zSHPgbkcIedBmG+Sk8Td5IZMAdi7Ufhmg9Wl7/jup/WrFbf5jh3MgKfgq
l4FKIOELgvi9T+ESaOuG2lCDL7sBzGqTdaaUn2L99KxOLABqHIgZ3JHUUNbVbOlP
mcIXa/7FVAza9bcc4JrY/gzQR1ztPySLtkG6idzdWIwJzwrK5dDftq5VVTrOcfU2
b8yGKZIUa9/pkpVGb8A0+0yeevDonYhnhi0+HN8EnV9Y1iDkWoj/6Jn8cT+DFjE+
jBv1HIQWLvLw5agP6P7xzU8WV+r2cZwQ63D8tayfC8baedJp9Dz0G5LFbnF/cSvc
/gqpXr+NAgMBAAECggEAWx82EAQrLhjCdBnhBCNVzqQiKpuu73AtZrAy20IixRV5
j7RF95oDnamTxkKcCXlyIggPMRJi74Uke2rMaCsUZw3fWgENAleTPkiR1QiNvxHG
IBhrNLgAWp5ncR8Uqw0Dt8QfGUF/3aDwsOyhZ9Nbr/o/gIqpkfkC7xVHFHdCjcp9
StvzhNeZIq1U/NYULrtnwkOnvV2f2dUEblruP6SzLksr4ZZXuSSC5RmX/U53GP0z
zOJj5pR2AJzjYjE4NuFxn4xtgj4eYSoGK0evyij01HFFHxjLQtLbeGo19WIXmJjA
0BwzJm7FuwbrhPzd49O66isRATXoA95/DAmQzha+OQKBgQD9DHQZ/DAvM1Xyw60v
LyPdj2+KZFPeU9P/bNzf+ZiJduKyRZWrFpnGpvyuSY3lje39HrUDJfpozO34H4BW
mrwe6uHfkDgAAvDaHzpt0mZ69XTCpGEjTtFQBXR+CIHJoAv8o11KewJNoOScAw/l
ElyT8S+ZCb6ZKWOGzRClNa7stwKBgQDDIGU1k/VRs/HVZh8SQSh42QWwdZU4PFVy
ITjcnqqCag4c6n9PXF3N846f/pXtYNW0HLJEmBcc7boea/sq18CCHiulLzlTWaVS
LwejTjZRHaJYxud8aBN7CbwZG0N6MsqMnnHtzFBEXu9m73zSKsgr/0A4Acbz67QX
tGSu06a52wKBgHkvr6KKLiFMuoqqv5PrRYfkG4zxg2DkUJDw986j4DNlJiguPwFS
r459hmGJhFU9ZY5lWFcLpyLtkcHUhEf1jsZXwpionskSn3o2nmrd6opUZviYdJTO
OFvUYPfC5zVCWrtBGXqD8pRuy00UAla4NnH7fcoS6p67PZjfOGuGjCF9AoGAWkFf
zzqTHKmpUNYdxSnSeKOZ2BdrYEm4FER9sr7Ji+1WfdWR8bl9wkfITwVJgDVsZBVp
+ASJnF3x2ySDVzvY1dbyxUNktsMejzclx0nkIf0dHQdUB910NVM5aDuOKLXZrtWT
STVaY2WuQuS/zc7wLDmzELTxu93ovZY5hAxucEUCgYBp08momp1VsxY+UVwRt/in
on1VquDXH0oinGHn9tMGNrvTb0M7u1UxXldP3KiIP35QZbumZK/kiFod3p6fJ+OU
nzvOGfUJga8KRGJAAenaKpxCw4S9RASrDoilCtlWDM4dBneB4daj4NoT0WNkSmCY
/kRDA9tgjNCiPbGaoIwTiw==
-----END PRIVATE KEY-----
`
	crtFile, err := os.Create(filepath.Join(nginxCert, "utm.crt"))
	check(err)
	defer crtFile.Close()
	crtFile.WriteString(cert)

	keyFile, err := os.Create(filepath.Join(nginxCert, "utm.key"))
	check(err)
	defer keyFile.Close()
	keyFile.WriteString(key)
}


func initializeElastic(secret string) {
	// wait for elastic to be ready
	baseURL := "http://127.0.0.1:9200/"
	log.Println("Waiting for the search engine")
	for {
		time.Sleep(50 * time.Second)

		_, err := grequests.Get(baseURL + "_cluster/healt", &grequests.RequestOptions{
			Params: map[string]string{
				"wait_for_status": "green",
				"timeout":         "50s",
			},
		})

		if err == nil {
			break
		}

		log.Println("Search engine is taking a long time to get ready, please wait.")
	}

	// configure elastic
	indexPrefix := "index-" + secret
	initialIndex := indexPrefix + "-000001"
	// create alias
	_, err := grequests.Post(baseURL + "_aliases", &grequests.RequestOptions{
		JSON: map[string][]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"add": map[string]string{
						"index": initialIndex,
						"alias": indexPrefix,
					},
				},
			},
		},
	})
	check(err)

	// create main index template
	_, err = grequests.Put(baseURL + "_template/main_index", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"index_patterns": []string{"index-*"},
			"settings": map[string]interface{}{
				"index.mapping.total_fields.limit": 10000,
				"opendistro.index_state_management.policy_id": "main_index_policy",
				"opendistro.index_state_management.rollover_alias": indexPrefix,
				"number_of_shards": 3,
				"number_of_replicas": 0,
			},
		},
	})
	check(err)

	// create template for generic index
	_, err = grequests.Put(baseURL + "_template/generic_index", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"index_patterns": []string{"generic-*"},
			"settings": map[string]interface{}{
				"index.mapping.total_fields.limit": 10000,
				"number_of_shards": 1,
				"number_of_replicas": 0,
			},
		},
	})
	check(err)

	// create templates
	for _, e := range []string{"dc", "utmstack", "utm"} {
		_, err = grequests.Put(baseURL + "_template/" + e + "_index", &grequests.RequestOptions{
			JSON: map[string]interface{}{
				"index_patterns": []string{e + "-*"},
				"settings": map[string]interface{}{
					"number_of_shards": 1,
					"number_of_replicas": 0,
				},
			},
		})
		check(err)
	}

	// enable snapshots
	_, err = grequests.Put(baseURL + "_snapshot/main_index", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"type": "fs",
			"settings": map[string]interface{}{
				"location": "backups",
			},
		},
	})
	check(err)

	_, err = grequests.Put(baseURL + "_snapshot/utm_geoip", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"type": "fs",
			"settings": map[string]interface{}{
				"location": "utm-geoip",
			},
		},
	})
	check(err)

	// create ISM policy
	_, err = grequests.Put(baseURL + "_opendistro/_ism/policies/main_index_policy", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"policy": map[string]interface{}{
				"description": "Main Index Lifecycle",
				"default_state": "ingest",
				"states": []interface{}{
					map[string]interface{}{
						"name": "ingest",
						"actions": []interface{}{
							map[string]interface{}{
								"rollover": map[string]interface{}{
									"min_doc_count": 30000000,
									"min_size": "15gb",
								},
							},
						},
						"transitions": []interface{}{
							map[string]string{
								"state_name": "search",
							},
						},
					},
					map[string]interface{}{
						"name": "search",
						"actions": []interface{}{
							map[string]interface{}{
								"snapshot": map[string]string{
									"repository": "main_index",
									"snapshot": "incremental",
								},
							},
						},
						"transitions": []interface{}{
							map[string]interface{}{
								"state_name": "read",
								"conditions": map[string]string{
									"min_index_age": "30d",
								},
							},
						},
					},
					map[string]interface{}{
						"name": "read",
						"actions": []interface{}{
							map[string]interface{}{
								"force_merge": map[string]interface{}{
									"max_num_segments": 1,
								},
							},
							map[string]interface{}{
								"snapshot": map[string]interface{}{
									"repository": "main_index",
									"snapshot": "incremental",
								},
							},
						},
						"transitions": []interface{}{},
					},
				},
			},
		},
	})
	check(err)

	// create initial index
	_, err = grequests.Put(baseURL + initialIndex, &grequests.RequestOptions{
		JSON: map[string]interface{}{},
	})
	check(err)
}

func initializePostgres(dbPassword string, clientName string, clientDomain string, 
	clientPrefix string, clientMail string) {
	// Connecting to PostgreSQL
	psqlconn := fmt.Sprintf("host=localhost port=5432 user=postgres password=%s sslmode=disable",
		dbPassword)
	srv, err := sql.Open("postgres", psqlconn)
	check(err)
	
	// Close connection when finish
	defer srv.Close()

	// Check connection status
	err = srv.Ping()
	check(err)

	// Crating utmstack
	_, err = srv.Exec("CREATE DATABASE utmstack")
	check(err)

	// Connecting to utmstack
	psqlconn = fmt.Sprintf("host=localhost port=5432 user=postgres password=%s sslmode=disable database=utmstack",
		dbPassword)
	db, err := sql.Open("postgres", psqlconn)
	check(err)
	
	// Close connection when finish
	defer db.Close()

	// Check connection status
	err = db.Ping()
	check(err)

	// Creating utm_client
	_, err = db.Exec(`CREATE TABLE public.utm_client (		
	id serial NOT NULL,
	client_name varchar(100) NULL,
	client_domain varchar(100) NULL,
	client_prefix varchar(10) NULL,
	client_mail varchar(100) NULL,
	client_user varchar(50) NULL,
	client_pass varchar(50) NULL,
	client_licence_creation timestamp(0) NULL,
	client_licence_expire timestamp(0) NULL,
	client_licence_id varchar(100) NULL,
	client_licence_verified bool NOT NULL,
	CONSTRAINT utm_client_pkey PRIMARY KEY (id)
	);`)
	check(err)

	// Insert client data
	_, err = db.Exec(`INSERT INTO public.utm_client (
	client_name, client_domain, client_prefix, 
	client_mail, client_user, client_pass, client_licence_verified
	) VALUES ($1, $2, $3, $4, 'admin', $5, false);`, 
	clientName, clientDomain, clientPrefix, clientMail, dbPassword)
	check(err)
}
