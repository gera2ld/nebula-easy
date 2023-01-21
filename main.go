package main

import (
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var nebulaCert = readEnv("NEBULA_CERT", "nebula-cert")
var dataPath = readEnv("DATA_PATH", "data/db.json")

//go:embed all:dist/*
var staticFs embed.FS

type NebulaCA struct {
	Name string `json:"name"`
	Crt  string `json:"crt"`
}

type NebulaSecretsCA struct {
	Key string `json:"key"`
}

type NebulaSecrets struct {
	Ca NebulaSecretsCA `json:"ca"`
}

type NebulaHost struct {
	HostType     string `json:"type"`
	Name         string `json:"name"`
	Ip           string `json:"ip"`
	Relay        bool   `json:"relay"`
	PublicIpPort string `json:"publicIpPort"`
}

type NebulaNetwork struct {
	Name    string       `json:"name"`
	IpRange string       `json:"ipRange"`
	Hosts   []NebulaHost `json:"hosts"`
}

type NebulaData struct {
	Ca       NebulaCA        `json:"ca"`
	Secrets  NebulaSecrets   `json:"secrets"`
	Networks []NebulaNetwork `json:"networks"`
}

type SpaFileSystem struct {
	root http.FileSystem
}

type SignCertParams struct {
	Name    string `json:"name"`
	IpRange string `json:"ipRange"`
	Pub     string `json:"pub"`
}

var nebulaData NebulaData

func readEnv(key string, def string) string {
	value := os.Getenv(key)
	if value == "" {
		value = def
	}
	return value
}

func (fs *SpaFileSystem) Open(name string) (http.File, error) {
	log.Println("Open file:", name)
	f, err := fs.root.Open(name)
	if os.IsNotExist(err) {
		return fs.root.Open("/index.html")
	}
	return f, err
}

func sendResponse(w http.ResponseWriter, result interface{}, err error, status int) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)
	data := make(map[string]interface{})
	data["result"] = result
	if err != nil {
		data["error"] = err.Error()
	}
	body, _ := json.Marshal(data)
	w.Write(body)
}

func readRequestBody(w http.ResponseWriter, r *http.Request, data any) error {
	err := json.NewDecoder(r.Body).Decode(data)
	if err != nil {
		sendResponse(w, nil, err, http.StatusBadRequest)
		return err
	}
	return nil
}

func handleApi(w http.ResponseWriter, r *http.Request) {
	command := strings.Split(r.URL.Path, "/")[2]
	log.Println("API:", command)
	if r.Method != "POST" {
		sendResponse(w, nil, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

	if command == "loadData" {
		data := make(map[string]any)
		data["ca"] = nebulaData.Ca
		data["networks"] = nebulaData.Networks
		sendResponse(w, data, nil, http.StatusOK)
	} else if command == "dumpData" {
		var data struct {
			Networks []NebulaNetwork `json:"networks"`
		}
		if readRequestBody(w, r, &data) != nil {
			return
		}
		dumpData(data.Networks)
		sendResponse(w, nil, nil, http.StatusOK)
	} else if command == "createCA" {
		var data string
		if readRequestBody(w, r, &data) != nil {
			return
		}
		result, err := createCA(data)
		sendResponse(w, result, err, http.StatusOK)
	} else if command == "signCert" {
		var data SignCertParams
		if readRequestBody(w, r, &data) != nil {
			return
		}
		log.Printf("%+v\n", data)
		result, err := signCert(data)
		sendResponse(w, result, err, http.StatusOK)
	} else if command == "getLighthouseConfig" {
		result, err := getLighthouseConfig()
		sendResponse(w, result, err, http.StatusOK)
	} else if command == "getHostConfig" {
		var data struct {
			staticHostMap   map[string][]string
			lighthouseHosts []string
			amRelay         bool
			relays          []string
		}
		if readRequestBody(w, r, &data) != nil {
			return
		}
		result, err := getHostConfig(data)
		sendResponse(w, result, err, http.StatusOK)
	}
}

func main() {
	loadData()
	http.HandleFunc("/api/", handleApi)
	fsDist, err := fs.Sub(staticFs, "dist")
	if err != nil {
		log.Fatal(err)
	}
	fs := http.FileServer(&SpaFileSystem{http.FS(fsDist)})
	http.Handle("/", fs)
	http.ListenAndServe(":4000", nil)
}

func RunCommandSafely(name string, args []string, cwd string) error {
	log.Println(name, args)
	cmd := exec.Command(name, args...)
	cmd.Dir = cwd
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func loadData() error {
	bytes, err := os.ReadFile(dataPath)
	if err == nil {
		err = json.Unmarshal([]byte(bytes), &nebulaData)
	}
	return err
}

func dumpData(data []NebulaNetwork) {
	log.Println("> Dump data")
	if data != nil {
		nebulaData.Networks = data
	}
	bytes, _ := json.Marshal(nebulaData)
	os.MkdirAll(filepath.Dir(dataPath), 0750)
	os.WriteFile(dataPath, bytes, 0644)
}

func createCA(name string) (map[string]interface{}, error) {
	cwd, _ := os.MkdirTemp("", "nebula-easy")
	defer os.RemoveAll(cwd)
	err := RunCommandSafely(nebulaCert, []string{"ca", "-name", name}, cwd)
	if err != nil {
		log.Println("Failed running nebula-cert ca")
		return nil, err
	}
	key, err := os.ReadFile(cwd + "/ca.key")
	if err != nil {
		log.Println("Failed reading ca.key")
		return nil, err
	}
	crt, err := os.ReadFile(cwd + "/ca.crt")
	if err != nil {
		log.Println("Failed reading ca.crt")
		return nil, err
	}
	nebulaData.Ca.Name = name
	nebulaData.Ca.Crt = string(crt)
	nebulaData.Secrets.Ca.Key = string(key)
	dumpData(nil)
	data := map[string]interface{}{"crt": string(crt)}
	return data, nil
}

func signCert(params SignCertParams) (map[string]interface{}, error) {
	if nebulaData.Secrets.Ca.Key == "" {
		return nil, errors.New("CA not found")
	}
	cwd, _ := os.MkdirTemp("", "nebula-easy")
	defer os.RemoveAll(cwd)
	os.WriteFile(cwd+"/ca.key", []byte(nebulaData.Secrets.Ca.Key), 0644)
	os.WriteFile(cwd+"/ca.crt", []byte(nebulaData.Ca.Crt), 0644)
	if params.Pub != "" {
		os.WriteFile(cwd+"/host.pub", []byte(params.Pub), 0644)
	}
	args := []string{"sign", "-name", params.Name, "-ip", params.IpRange, "-out-crt", "host.crt"}
	if params.Pub != "" {
		args = append(args, "-in-pub", "host.pub")
	} else {
		args = append(args, "-out-key", "host.key")
	}
	err := RunCommandSafely(nebulaCert, args, cwd)
	if err != nil {
		log.Println("Failed running nebula-cert sign")
		return nil, err
	}
	crt, err := os.ReadFile(cwd + "/host.crt")
	if err != nil {
		log.Println("Failed reading host.crt")
		return nil, err
	}
	data := map[string]interface{}{"crt": string(crt)}
	if params.Pub == "" {
		key, err := os.ReadFile(cwd + "/host.key")
		if err != nil {
			log.Println("Failed reading host.key")
			return nil, err
		}
		data["key"] = string(key)
	}
	return data, nil
}

func getBaseConfig() map[string]any {
	return map[string]any{
		"pki": map[string]string{
			"ca":   "/etc/nebula/ca.crt",
			"cert": "/etc/nebula/host.crt",
			"key":  "/etc/nebula/host.key",
		},
		"lighthouse": map[string]any{
			"am_lighthouse": true,
		},
		"listen": map[string]any{
			"host": "0.0.0.0",
			"port": 4242,
		},
	}
}

func getRelayConfig(amRelay bool, relays []string) map[string]any {
	if amRelay {
		return map[string]any{
			"am_relay":   true,
			"use_relays": false,
		}
	}
	if relays == nil {
		relays = make([]string, 0)
	}
	return map[string]any{
		"am_relay":   false,
		"use_relays": true,
		"relays":     relays,
	}
}

func getLighthouseConfig() (map[string]any, error) {
	config := getBaseConfig()
	config["relay"] = getRelayConfig(true, nil)
	return config, nil
}

func getHostConfig(params struct {
	staticHostMap   map[string][]string
	lighthouseHosts []string
	amRelay         bool
	relays          []string
}) (map[string]any, error) {
	config := getBaseConfig()
	config["lighthouse"] = map[string]any{
		"am_lighthouse": false,
		"hosts":         params.lighthouseHosts,
	}
	config["static_host_map"] = params.staticHostMap
	config["relay"] = getRelayConfig(params.amRelay, params.relays)
	return config, nil
}
