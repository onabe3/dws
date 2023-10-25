package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// tarBuildContext: Dockerfileをtarアーカイブに変換する関数
func tarBuildContext(dockerfilePath string, dockerfileName string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Dockerfileの内容を読み込む
	dockerfile, err := os.Open(dockerfilePath)
	if err != nil {
		return nil, err
	}
	defer dockerfile.Close()

	dockerfileContent, err := io.ReadAll(dockerfile)
	if err != nil {
		return nil, err
	}

	// tarにDockerfileを追加する
	tarHeader := &tar.Header{
		Name: dockerfileName,
		Size: int64(len(dockerfileContent)),
	}
	if err := tw.WriteHeader(tarHeader); err != nil {
		return nil, err
	}
	if _, err := tw.Write(dockerfileContent); err != nil {
		return nil, err
	}

	return buf, nil
}

func main() {
	// Docker clientの初期化
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Initialize Docker client error: %v", err)
	}

	// Dockerfileの内容を定義
	dockerfileContent := `
FROM ubuntu:latest
# 追加の設定やインストールコマンド
`
	// Dockerfileに内容を書き込む
	dockerfile, err := os.Create("Dockerfile")
	if err != nil {
		log.Fatalf("Failed to create Dockerfile: %v", err)
	}
	_, err = dockerfile.WriteString(dockerfileContent)
	if err != nil {
		log.Fatalf("Failed to write Dockerfile: %v", err)
	}
	dockerfile.Close()

	// Dockerfileをtarアーカイブに変換
	buildContext, err := tarBuildContext("Dockerfile", "Dockerfile")
	if err != nil {
		log.Fatalf("Failed to create tar build context: %v", err)
	}

	// イメージ名を定義
	imageName := "my-ubuntu-image"

	// Dockerイメージをビルドする
	buildResp, err := cli.ImageBuild(
		context.Background(),
		io.NopCloser(buildContext),
		types.ImageBuildOptions{
			Tags:       []string{imageName},
			Dockerfile: "Dockerfile",
		},
	)
	if err != nil {
		log.Fatalf("Image build error: %v", err)
	}
	defer buildResp.Body.Close()

	// Buildのログを表示する
	fmt.Println((stdcopy.StdCopy(os.Stdout, os.Stderr, buildResp.Body)))

	// イメージを元に新しいコンテナを作成する
	config := &container.Config{
		Image: imageName,
	}

	// コンテナ名を定義
	containerName := "ASANO"

	resp, err := cli.ContainerCreate(context.Background(), config, nil, nil, nil, containerName)
	if err != nil {
		log.Fatalf("Container create error: %v", err)
	}

	// コンテナを起動する
	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatalf("Container start error: %v", err)
	}

	fmt.Printf("Container %s started\n", resp.ID)
}
