package main

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	//go:embed index.json
	index string

	//go:embed manifest_amd64.json
	manifestAmd64 string

	//go:embed manifest_arm64.json
	manifestArm64 string

	//go:embed config_amd64.json
	configAmd64 string

	//go:embed config_arm64.json
	configArm64 string

	//go:embed layer_amd64.tar.gz
	layerAmd64 string

	//go:embed layer_arm64.tar.gz
	layerArm64 string
)

func getV2(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got request to %s\n", r.URL.Path)

	_, err := io.WriteString(w, "{}")
	if err != nil {
		fmt.Printf("error writing response: %v\n", err)
	}
}

func setHandleFunc(mux *http.ServeMux, path string, contentType string, content string) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("got request to %s\n", r.URL.Path)

		w.Header().Add("Content-Type", contentType)
		_, err := io.WriteString(w, content)
		if err != nil {
			fmt.Printf("error writing response: %v\n", err)
		}
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2", getV2)
	mux.HandleFunc("/v2/", getV2)
	setHandleFunc(mux, "/v2/uniget-org/tools/jq/manifests/latest", "application/vnd.oci.image.index.v1+json", index)
	setHandleFunc(mux, "/v2/uniget-org/tools/jq/manifests/sha256:be2ea365585029e1fe8b5d1b871f5da3b3f1b228afea8e78d4833cf92138d695", "application/vnd.oci.image.manifest.v1+json", manifestAmd64)
	setHandleFunc(mux, "/v2/uniget-org/tools/jq/manifests/sha256:9fdf2c55f9079be137c4c03ba8038f89319f0964c098f48fc57f35645b9ecb0f", "application/vnd.oci.image.manifest.v1+json", manifestArm64)
	setHandleFunc(mux, "/v2/uniget-org/tools/jq/blobs/sha256:c80aebd6980266c68ccff639dcabf8ba089f20ee5917541aacf02d0f002d4f5a", "application/vnd.oci.image.config.v1+json", configAmd64)
	setHandleFunc(mux, "/v2/uniget-org/tools/jq/blobs/sha256:ad7bb1709f30ce450621a8e334393ddb52996a4938e367b1889441db0cf8bb6f", "application/vnd.oci.image.config.v1+json", configArm64)
	setHandleFunc(mux, "/v2/uniget-org/tools/jq/blobs/sha256:8f6a01445b829e82a8e6d5a66d646c1d884d0917df2c9a415a194cf273ac189d", "application/vnd.oci.image.layer.v1.tar+gzip", layerAmd64)
	setHandleFunc(mux, "/v2/uniget-org/tools/jq/blobs/sha256:63986aba908f3edf7fab57738f5126e64e566981a33d370097377e597b74384a", "application/vnd.oci.image.layer.v1.tar+gzip", layerArm64)

	server := &http.Server{
		Addr:         ":5000",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
