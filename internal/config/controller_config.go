package config

import (
	"fmt"

	"io/ioutil"
	"os"

	"github.com/asaskevich/govalidator"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger"
	defaults "github.com/mcuadros/go-defaults"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

// ControllerConfig provide helm broker configuration
// Supported tags:
//	- json: 		github.com/ghodss/yaml
//	- envconfig: 	github.com/vrischmann/envconfig
//	- default: 		github.com/mcuadros/go-defaults
//	- valid         github.com/asaskevich/govalidator
// Example of valid tag: `valid:"alphanum,required"`
// Combining many tags: tags have to be separated by WHITESPACE: `json:"port" default:"8080" valid:"required"`
type ControllerConfig struct {
	TmpDir                   string
	Logger                   logger.Config
	KubeconfigPath           string `envconfig:"optional"`
	Namespace                string
	ServiceName              string
	ClusterServiceBrokerName string
	Storage                  []storage.Config `valid:"required"`
	DevelopMode              bool
	UploadServiceURL         string `default:"http://assetstore-asset-upload-service.kyma-system.svc.cluster.local:3000"`
}

// LoadControllerConfig method has following strategy:
// 1. Check env variable 'APP_CONFIG_FILE_NAME', if exists load configuration from specified file
// 2. Read configuration from environment variables (will override configuration from file)
// 3. Apply defaults
// 4. Validate
func LoadControllerConfig(verbose bool) (*ControllerConfig, error) {
	outCfg := ControllerConfig{}

	cfgFile := os.Getenv("APP_CONFIG_FILE_NAME")
	if cfgFile != "" {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			return nil, errors.Wrapf(err, "while opening config file [%s]", cfgFile)
		}
		fileConfig := ControllerConfig{}
		if err := yaml.Unmarshal(b, &fileConfig); err != nil {
			return nil, errors.Wrap(err, "while unmarshalling config from file")
		}
		outCfg = fileConfig
		// fmt.Printf used, because logger will be created after reading configuration
		if verbose {
			fmt.Printf("Config after applying values from file: %+v\n", outCfg)
		}
	}

	envConf := ControllerConfig{}
	if err := envconfig.InitWithOptions(&envConf, envconfig.Options{Prefix: "APP", AllOptional: true, AllowUnexported: true}); err != nil {
		return nil, errors.Wrap(err, "while reading configuration from environment variables")
	}

	if err := mergo.MergeWithOverwrite(&outCfg, &envConf); err != nil {
		return nil, errors.Wrap(err, "while merging config from environment variables")
	}
	if verbose {
		fmt.Printf("Config after applying values from environment variables: %+v\n", outCfg)
	}

	defaults.SetDefaults(&outCfg)

	if verbose {
		fmt.Printf("Config after applying defaults: %+v\n", outCfg)
	}
	if _, err := govalidator.ValidateStruct(outCfg); err != nil {
		return nil, errors.Wrap(err, "while validating configuration object")
	}
	return &outCfg, nil
}
