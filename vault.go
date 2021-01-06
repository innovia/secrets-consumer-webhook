package main

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type vault struct {
	config struct {
		enabled                        bool
		addr                           string
		tlsSecretName                  string
		vaultCACert                    string
		VaultCAKey                     string
		path                           string
		role                           string
		tokenPath                      string
		authPath                       string
		backend                        string
		useSecretNamesAsKeys           bool
		gcpServiceAccountKeySecretName string
		version                        string
	}
}

func (vault *vault) mutateContainer(container corev1.Container) corev1.Container {
	envVars := vault.setEnvVars()
	container.Env = append(container.Env, envVars...)

	if vault.config.useSecretNamesAsKeys {
		container.Env = append(container.Env, []corev1.EnvVar{
			{
				Name:  "VAULT_USE_SECRET_NAMES_AS_KEYS",
				Value: "true",
			},
		}...)
	}

	if vault.config.version != "" {
		container.Env = append(container.Env, []corev1.EnvVar{
			{
				Name:  "VAULT_SECRET_VERSION",
				Value: vault.config.version,
			},
		}...)
	}

	// Mount google service account key if given
	if vault.config.gcpServiceAccountKeySecretName != "" {
		container.VolumeMounts = append(container.VolumeMounts, []corev1.VolumeMount{
			{
				Name:      "google-cloud-key",
				MountPath: VolumeMountGoogleCloudKeyPath,
			},
		}...)
	}

	if vault.config.tlsSecretName != "" {
		mountPath := fmt.Sprintf("%s/%s", VaultTLSMountPath, vault.config.vaultCACert)
		volumeName := VaultTLSVolumeName

		container.Env = append(container.Env, []corev1.EnvVar{
			{
				Name:  "VAULT_CACERT",
				Value: mountPath,
			},
		}...)
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
			SubPath:   vault.config.vaultCACert,
		})
	} else {
		container.Env = append(container.Env, []corev1.EnvVar{
			{
				Name:  "VAULT_SKIP_VERIFY",
				Value: "true",
			},
		}...)
	}
	return container
}

func (vault *vault) setEnvVars() []corev1.EnvVar {
	var envVars []corev1.EnvVar
	envVars = append(envVars, []corev1.EnvVar{
		{
			Name:  "VAULT_ADDR",
			Value: vault.config.addr,
		},
		{
			Name:  "VAULT_PATH",
			Value: vault.config.path,
		}, {
			Name:  "VAULT_ROLE",
			Value: vault.config.role,
		},
	}...)

	if vault.config.backend == "gcp" {
		envVars = append(envVars, []corev1.EnvVar{
			{
				Name:  "VAULT_BACKEND",
				Value: "gcp",
			},
		}...)

		if vault.config.gcpServiceAccountKeySecretName != "" {
			envVars = append(envVars, []corev1.EnvVar{
				{
					Name:  "GOOGLE_APPLICATION_CREDENTIALS",
					Value: "/var/run/secret/cloud.google.com/service-account.json",
				},
			}...)
		}
	}

	if vault.config.tokenPath != "" {
		envVars = append(envVars, []corev1.EnvVar{
			{
				Name:  "TOKEN_PATH",
				Value: vault.config.tokenPath,
			},
		}...)
	}

	return envVars
}
