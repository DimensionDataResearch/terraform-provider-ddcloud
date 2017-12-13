package ddcloud

import (
	"bytes"
	"encoding/base64"
	"encoding/pem"
	"io/ioutil"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"golang.org/x/crypto/pkcs12"
)

const (
	resourceKeyPFXFile        = "file"
	resourceKeyPFXPassword    = "password"
	resourceKeyPFXCertificate = "certificate"
	resourceKeyPFXPrivateKey  = "private_key"
)

var pemBase64 = base64.StdEncoding

func dataSourcePFX() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePFXRead,

		Schema: map[string]*schema.Schema{
			resourceKeyPFXFile: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the PFX file",
			},
			resourceKeyPFXPassword: &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The password for the PFX file",
			},
			resourceKeyPFXCertificate: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The (first) certificate in the PFX file",
			},
			resourceKeyPFXPrivateKey: &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The (first) private key in the PFX file",
			},
		},
	}
}

// Read a network domain data source.
func dataSourcePFXRead(data *schema.ResourceData, provider interface{}) error {
	fileName := data.Get(resourceKeyPFXFile).(string)

	pfxData, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("Failed to read PFX data from '%s': %s", fileName, err.Error())

		return err
	}

	log.Printf("Read PFX data from '%s'.", fileName)

	pfxPassword := data.Get(resourceKeyPFXPassword).(string)
	pemBlocks, err := pkcs12.ToPEM(pfxData, pfxPassword)
	if err != nil {
		log.Printf("Failed to decode PFX data from '%s': %s", fileName, err.Error())

		return err
	}

	var (
		certificatePEM string
		privateKeyPEM  string
	)
	for _, pemBlock := range pemBlocks {
		switch pemBlock.Type {
		case "CERTIFICATE":
			if certificatePEM == "" {
				certificatePEM, err = pemToString(pemBlock)
				if err != nil {
					return err
				}
			}
		case "PRIVATE KEY":
			if privateKeyPEM == "" {
				privateKeyPEM, err = pemToString(pemBlock)
				if err != nil {
					return err
				}
			}
		}
	}

	data.Set(resourceKeyPFXCertificate, certificatePEM)
	data.Set(resourceKeyPFXPrivateKey, privateKeyPEM)

	return nil
}

func pemToString(pemBlock *pem.Block) (string, error) {
	var buffer bytes.Buffer
	err := pem.Encode(&buffer, pemBlock)
	if err != nil {
		log.Printf("Failed to decode '%s' PEM block: %s", pemBlock.Type, err.Error())

		return "", err
	}

	return buffer.String(), nil
}
