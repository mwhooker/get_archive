package main

import (
	"flag"
	"fmt"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	"log"
	"os"
	"strings"
)

/**
Downloader let's us download latest or pegged version of a project

a project must follow our s3 scheme

bucket:
	s3://wercker-{environment}/

A file whose content is the "short-rev" of a git commit
	/{project}/{branch}/HEAD

The project file for a given commit on a given branch.
Branch is redundant with commit, but we lay it out like this to make it easier on us humans.
	/{project}/{branch}/{project}.{short-rev}.tgz
*/
type Downloader struct {
	Bucket  *s3.Bucket
	branch  string
	project string
}

func NewDownloader(auth *aws.Auth, project string, env string, branch string) *Downloader {
	s3_conn := s3.New(*auth, aws.USEast)
	bucket_name := fmt.Sprintf("wercker-%s", env)
	bucket := s3_conn.Bucket(bucket_name)
	return &Downloader{bucket, branch, project}
}

func (dl *Downloader) getHead() (string, error) {
	path := fmt.Sprintf("/%s/%s/HEAD", dl.project, dl.branch)
	log.Println("getting HEAD from", path)
	data, err := dl.Bucket.Get(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data[:])), nil
}

func (dl *Downloader) GetLatest(destination string) error {
	// s3://wercker-development/kiddie-pool/master/
	//wercker-development/kiddie-pool/docker/kiddie-pool.7d3bb37.tgz
	head, err := dl.getHead()
	if err != nil {
		panic(err)
	}
	path := fmt.Sprintf("/%s/%s/%s.%s.tgz", dl.project, dl.branch, dl.project, head)
	log.Println(path)
	data, err := dl.Bucket.Get(path)
	if err != nil {
		panic(err)
	}
	fp, err := os.Create(destination)
	if err != nil {
		panic(err)
	}
	_, err = fp.Write(data)
	if err != nil {
		panic(err)
	}
	err = fp.Chmod(0544)
	if err != nil {
		panic(err)
	}

	return nil
}

func main() {
	var environment = flag.String("environment", "development", "Which environment to fetch the build from (development, production, etc).")
	var branch = flag.String("branch", "master", "Which branch of the build we want.")
	// Project must follow our s3 scheme (see above)
	var projectName = flag.String("project", "kiddie-pool", "Project we want to download.")
	var destination = flag.String("destination", "", "The name to write the file to. If empty, use the name of the file in S3.")
	var accessKey = flag.String("access-key", "", "AWS access key. Leave empty to get from environment.")
	var secretKey = flag.String("secret-key", "", "AWS secret access key. Leave empty to get from environment.")
	flag.Parse()
	var auth aws.Auth
	var err error
	if len(*accessKey) > 0 && len(*secretKey) > 0 {
		log.Println("using provided aws keys", *accessKey)
		auth = aws.Auth{AccessKey: *accessKey, SecretKey: *secretKey}
	} else {
		log.Println("using aws keys from environment", *accessKey)
		auth, err = aws.EnvAuth()
		if err != nil {
			panic(err)
		}
	}

	dl := NewDownloader(&auth, *projectName, *environment, *branch)
	if len(*destination) > 0 {
		dl.GetLatest(*destination)
	} else {
		dl.GetLatest(*projectName)
	}
}
