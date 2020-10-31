package seed

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mitchellh/mapstructure"
	"github.com/sgtcodfish/scrimplb/constants"
	"github.com/sgtcodfish/scrimplb/resolver"
)

// S3Provider can retrieve a seed from an object in an S3 bucket.
// Required permissions for a load balancer are: GetObject, PutObject, ListBucket
// Required permissions for an application server are: GetObject, ListBucket
// Without ListBucket you'll get an AccessDenied when you try to fetch a nonexistant seed
type S3Provider struct {
	Bucket string
	Region string
	Key    string

	// accessKey       string
	// secretAccessKey string
}

// NewS3Provider creates a new S3 seed provider from the given config.
// "availability-zone" is optional; if not given it will be deduced from instance
// metadata if running on EC2.
func NewS3Provider(config map[string]interface{}) (*S3Provider, error) {
	var provider S3Provider

	err := mapstructure.Decode(config, &provider)

	if err != nil {
		return nil, fmt.Errorf("couldn't parse s3 provider config: %w", err)
	}

	if provider.Bucket == "" {
		return nil, errors.New("missing required 'bucket' in provider config")
	}

	if provider.Region == "" {
		return nil, errors.New("missing required 'region' in provider config")
	}

	if provider.Key == "" {
		provider.Key = constants.DefaultKey
	}

	return &provider, nil
}

// FetchSeed makes a call to S3 to retrieve the seed for the AZ this
// instance is located in.
func (s *S3Provider) FetchSeed() (Seeds, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.Region)},
	)

	if err != nil {
		return Seeds{}, err
	}

	downloader := s3manager.NewDownloader(sess)

	buf := aws.NewWriteAtBuffer([]byte{})

	numBytes, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(s.Bucket),
			Key:    aws.String(s.Key),
		})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return Seeds{}, nil

			default:
				return Seeds{}, fmt.Errorf("unable to download seed: %w", err)
			}
		} else {
			return Seeds{}, fmt.Errorf("unable to download seed; non-AWS error: %w", err)
		}
	}

	log.Printf("downloaded %d bytes from S3 for seed\n", numBytes)

	var seeds Seeds

	err = json.Unmarshal(buf.Bytes(), &seeds)

	if err != nil {
		return Seeds{}, fmt.Errorf("unable to parse downloaded seed: %w", err)
	}

	return seeds, nil
}

// PushSeed fetches remote data, updates with local node details and then
// publishes the details to an S3 bucket so that future nodes can join
func (s *S3Provider) PushSeed(resolver resolver.IPResolver, port string) error {
	ip, err := resolver.ResolveIP()

	if err != nil {
		return err
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s.Region)},
	)

	if err != nil {
		return err
	}

	downloader := s3manager.NewDownloader(sess)

	buf := aws.NewWriteAtBuffer([]byte{})

	numBytes, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(s.Bucket),
			Key:    aws.String(s.Key),
		})

	var downloadedSeeds Seeds
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return fmt.Errorf("unable to push seed - no such bucket: %w", err)

			default:
				// ignore
			}
		}
	} else {
		log.Printf("downloaded %d bytes from S3 for seed\n", numBytes)
		err = json.Unmarshal(buf.Bytes(), &downloadedSeeds)

		if err != nil {
			log.Println("unable to parse downloaded seed, defaulting to overwriting file")
		}
	}

	seeds := Seeds{}
	if downloadedSeeds.Seeds != nil && len(downloadedSeeds.Seeds) > 0 {
		seeds.Seeds = downloadedSeeds.Seeds
		hasCurrent := false

		for _, s := range downloadedSeeds.Seeds {
			if s.Address == ip && s.Port == port {
				hasCurrent = true
				break
			}
		}

		if !hasCurrent {
			seeds.Seeds = append(seeds.Seeds, Seed{
				Address: ip,
				Port:    port,
			})
		} else {
			log.Println("skipping publishing seed as address is already published")
			return nil
		}
	} else {
		seeds.Seeds = []Seed{{
			Address: ip,
			Port:    port,
		}}
	}

	out, err := json.Marshal(seeds)

	if err != nil {
		return fmt.Errorf("couldn't marshal output for S3: %w", err)
	}

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
		Body:   bytes.NewReader(out),
	})

	if err != nil {
		return fmt.Errorf("couldn't upload S3 content: %w", err)
	}

	log.Println("successfully pushed seed to s3")

	return nil
}
