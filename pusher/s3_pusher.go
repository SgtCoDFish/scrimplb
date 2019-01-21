package pusher

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// S3Pusher pushes cluster join data to an AWS S3 Bucket
type S3Pusher struct {
	Bucket string
	Region string
}

// NewS3Pusher creates an S3Pusher from the given config, requiring at least
// a bucket and region
func NewS3Pusher(config map[string]interface{}) (*S3Pusher, error) {
	var pusher S3Pusher

	err := mapstructure.Decode(config, &pusher)

	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse s3 pusher config")
	}

	if pusher.Bucket == "" {
		return nil, errors.New("missing required 'bucket' in pusher config")
	}

	if pusher.Region == "" {
		return nil, errors.New("missing required 'region' in pusher config")
	}

	return &pusher, nil
}

// PushState connects to S3 and pushes the current state to an object
func (s *S3Pusher) PushState() error {
	return errors.New("nyi")
}
