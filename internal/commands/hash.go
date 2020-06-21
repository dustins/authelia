package commands

import (
	"fmt"
	"log"

	"github.com/simia-tech/crypt"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/configuration/schema"
)

func init() {
	HashPasswordCmd.Flags().BoolP("sha512", "z", false, fmt.Sprintf("use sha512 as the algorithm (changes iterations to %d, change with -i)", schema.DefaultPasswordSHA512Configuration.Iterations))
	HashPasswordCmd.Flags().IntP("iterations", "i", schema.DefaultPasswordConfiguration.Iterations, "set the number of hashing iterations")
	HashPasswordCmd.Flags().StringP("salt", "s", "", "set the salt string")
	HashPasswordCmd.Flags().IntP("memory", "m", schema.DefaultPasswordConfiguration.Memory, "[argon2id] set the amount of memory param (in MB)")
	HashPasswordCmd.Flags().IntP("parallelism", "p", schema.DefaultPasswordConfiguration.Parallelism, "[argon2id] set the parallelism param")
	HashPasswordCmd.Flags().IntP("key-length", "k", schema.DefaultPasswordConfiguration.KeyLength, "[argon2id] set the key length param")
	HashPasswordCmd.Flags().IntP("salt-length", "l", schema.DefaultPasswordConfiguration.SaltLength, "set the auto-generated salt length")
}

// HashPasswordCmd password hashing command.
var HashPasswordCmd = &cobra.Command{
	Use:   "hash-password [password]",
	Short: "Hash a password to be used in file-based users database. Default algorithm is argon2id.",
	Run: func(cobraCmd *cobra.Command, args []string) {
		sha512, _ := cobraCmd.Flags().GetBool("sha512")
		iterations, _ := cobraCmd.Flags().GetInt("iterations")
		salt, _ := cobraCmd.Flags().GetString("salt")
		keyLength, _ := cobraCmd.Flags().GetInt("key-length")
		saltLength, _ := cobraCmd.Flags().GetInt("salt-length")
		memory, _ := cobraCmd.Flags().GetInt("memory")
		parallelism, _ := cobraCmd.Flags().GetInt("parallelism")

		var err error
		var hash string
		var algorithm authentication.CryptAlgo

		if sha512 {
			if iterations == schema.DefaultPasswordConfiguration.Iterations {
				iterations = schema.DefaultPasswordSHA512Configuration.Iterations
			}
			algorithm = authentication.HashingAlgorithmSHA512
		} else {
			algorithm = authentication.HashingAlgorithmArgon2id
		}
		if salt != "" {
			salt = crypt.Base64Encoding.EncodeToString([]byte(salt))
		}

		hash, err = authentication.HashPassword(args[0], salt, algorithm, iterations, memory*1024, parallelism, keyLength, saltLength)
		if err != nil {
			log.Fatalf("Error occurred during hashing: %s\n", err)
		} else {
			fmt.Printf("Password hash: %s\n", hash)
		}
	},
	Args: cobra.MinimumNArgs(1),
}
