package cmd

import (
	"html/template"
	"os"

	"github.com/brocaar/lora-gateway-bridge/internal/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// when updating this template, don't forget to update config.md!
const configTemplate = `[general]
# debug=5, info=4, warning=3, error=2, fatal=1, panic=0
log_level = {{ .General.LogLevel }}


# Filters.
#
# These can be used to filter LoRaWAN frames to reduce bandwith usage between
# the gateway and LoRa Gateway Bride. Depending the used backend, filtering
# will be performed by the Packet Forwarder or LoRa Gateway Bridge.
[filters]

# NetIDs filters.
#
# The configured NetIDs will be used to filter uplink data frames.
# When left blank, no filtering will be performed on NetIDs.
#
# Example:
# net_ids=[
#   "000000",
#   "000001",
# ]
net_ids=[{{ range $index, $elm := .Filters.NetIDs }}
  "{{ $elm }}",{{ end }}
]

# JoinEUI filters.
#
# The configured JoinEUI ranges will be used to filter join-requests.
# When left blank, no filtering will be performed on JoinEUIs.
#
# Example:
# join_euis=[
#   ["0000000000000000", "00000000000000ff"],
#   ["000000000000ff00", "000000000000ffff"],
# ]
join_euis=[{{ range $index, $elm := .Filters.JoinEUIs }}
  ["{{ index $elm 0 }}", "{{ index $elm 1 }}"],{{ end }}
]


# Gateway backend configuration.
[backend]

# Backend type.
#
# Valid options are:
#   * semtech_udp
#   * basic_station
type="{{ .Backend.Type }}"


  # Semtech UDP packet-forwarder backend.
  [backend.semtech_udp]

  # ip:port to bind the UDP listener to
  #
  # Example: 0.0.0.0:1700 to listen on port 1700 for all network interfaces.
  # This is the listeren to which the packet-forwarder forwards its data
  # so make sure the 'serv_port_up' and 'serv_port_down' from your
  # packet-forwarder matches this port.
  udp_bind = "{{ .Backend.SemtechUDP.UDPBind }}"

  # Skip the CRC status-check of received packets
  #
  # This is only has effect when the packet-forwarder is configured to forward
  # LoRa frames with CRC errors.
  skip_crc_check = {{ .Backend.SemtechUDP.SkipCRCCheck }}

  # Fake RX timestamp.
  #
  # Fake the RX time when the gateway does not have GPS, in which case
  # the time would otherwise be unset.
  fake_rx_time={{ .Backend.SemtechUDP.FakeRxTime }}

{{ range $i, $config := .Backend.SemtechUDP.Configuration }}
    [[backend.semtech_udp.configuration]]
    gateway_id="{{ $config.GatewayID }}"
    base_file="{{ $config.BaseFile }}"
    output_file="{{ $config.OutputFile }}"
    restart_command="{{ $config.RestartCommand }}"
{{ end }}

  # Basic Station backend.
  [backend.basic_station]

  # ip:port to bind the Websocket listener to.
  bind="{{ .Backend.BasicStation.Bind }}"

  # TLS certificate and key files.
  #
  # When set, the websocket listener will use TLS to secure the connections
  # between the gateways and LoRa Gateway Bridge (optional).
  tls_cert="{{ .Backend.BasicStation.TLSCert }}"
  tls_key="{{ .Backend.BasicStation.TLSKey }}"

  # TLS CA certificate.
  #
  # When configured, LoRa Gateway Bridge will validate that the client
  # certificate of the gateway has been signed by this CA certificate.
  ca_cert="{{ .Backend.BasicStation.CACert }}"

  # Verify client vertificate CommonName
  #
  # Require that the CommonName on the client certificate matches
  # the EUI that the gateway is claiming to be.
  verify_cn={{ .Backend.BasicStation.VerifyCN }}

  # Ping interval.
  ping_interval="{{ .Backend.BasicStation.PingInterval }}"

  # Read timeout.
  #
  # This interval must be greater than the configured ping interval.
  read_timeout="{{ .Backend.BasicStation.ReadTimeout }}"

  # Write timeout.
  write_timeout="{{ .Backend.BasicStation.WriteTimeout }}"

  # Region.
  #
  # Please refer to the LoRaWAN Regional Parameters specification
  # for the complete list of common region names.
  region="{{ .Backend.BasicStation.Region }}"

  # Minimal frequency (Hz).
  frequency_min={{ .Backend.BasicStation.FrequencyMin }}

  # Maximum frequency (Hz).
  frequency_max={{ .Backend.BasicStation.FrequencyMax }}


# Integration configuration.
[integration]
# Payload marshaler.
#
# This defines how the MQTT payloads are encoded. Valid options are:
# * protobuf:  Protobuf encoding (this will become the LoRa Gateway Bridge v3 default)
# * json:      JSON encoding (easier for debugging, but less compact than 'protobuf')
marshaler="{{ .Integration.Marshaler }}"

  # MQTT integration configuration.
  [integration.mqtt]
  # Event topic template.
  event_topic_template="{{ .Integration.MQTT.EventTopicTemplate }}"

  # Command topic template.
  command_topic_template="{{ .Integration.MQTT.CommandTopicTemplate }}"

  # Maximum interval that will be waited between reconnection attempts when connection is lost.
  # Valid units are 'ms', 's', 'm', 'h'. Note that these values can be combined, e.g. '24h30m15s'.
  max_reconnect_interval="{{ .Integration.MQTT.MaxReconnectInterval }}"


  # MQTT authentication.
  [integration.mqtt.auth]
  # Type defines the MQTT authentication type to use.
  #
  # Set this to the name of one of the sections below.
  type="{{ .Integration.MQTT.Auth.Type }}"

    # Generic MQTT authentication.
    [integration.mqtt.auth.generic]
    # MQTT server (e.g. scheme://host:port where scheme is tcp, ssl or ws)
    server="{{ .Integration.MQTT.Auth.Generic.Server }}"

    # Connect with the given username (optional)
    username="{{ .Integration.MQTT.Auth.Generic.Username }}"

    # Connect with the given password (optional)
    password="{{ .Integration.MQTT.Auth.Generic.Password }}"

    # Quality of service level
    #
    # 0: at most once
    # 1: at least once
    # 2: exactly once
    #
    # Note: an increase of this value will decrease the performance.
    # For more information: https://www.hivemq.com/blog/mqtt-essentials-part-6-mqtt-quality-of-service-levels
    qos={{ .Integration.MQTT.Auth.Generic.QOS }}

    # Clean session
    #
    # Set the "clean session" flag in the connect message when this client
    # connects to an MQTT broker. By setting this flag you are indicating
    # that no messages saved by the broker for this client should be delivered.
    clean_session={{ .Integration.MQTT.Auth.Generic.CleanSession }}

    # Client ID
    #
    # Set the client id to be used by this client when connecting to the MQTT
    # broker. A client id must be no longer than 23 characters. When left blank,
    # a random id will be generated. This requires clean_session=true.
    client_id="{{ .Integration.MQTT.Auth.Generic.ClientID }}"

    # CA certificate file (optional)
    #
    # Use this when setting up a secure connection (when server uses ssl://...)
    # but the certificate used by the server is not trusted by any CA certificate
    # on the server (e.g. when self generated).
    ca_cert="{{ .Integration.MQTT.Auth.Generic.CACert }}"

    # mqtt TLS certificate file (optional)
    tls_cert="{{ .Integration.MQTT.Auth.Generic.TLSCert }}"

    # mqtt TLS key file (optional)
    tls_key="{{ .Integration.MQTT.Auth.Generic.TLSKey }}"


    # Google Cloud Platform Cloud IoT Core authentication.
    #
    # Please note that when using this authentication type, the MQTT topics
    # will be automatically set to match the MQTT topics as expected by
    # Cloud IoT Core.
    [integration.mqtt.auth.gcp_cloud_iot_core]
    # MQTT server.
    server="{{ .Integration.MQTT.Auth.GCPCloudIoTCore.Server }}"

    # Google Cloud IoT Core Device id.
    device_id="{{ .Integration.MQTT.Auth.GCPCloudIoTCore.DeviceID }}"

    # Google Cloud project id.
    project_id="{{ .Integration.MQTT.Auth.GCPCloudIoTCore.ProjectID }}"

    # Google Cloud region.
    cloud_region="{{ .Integration.MQTT.Auth.GCPCloudIoTCore.CloudRegion }}"

    # Google Cloud IoT registry id.
    registry_id="{{ .Integration.MQTT.Auth.GCPCloudIoTCore.RegistryID }}"

    # JWT token expiration time.
    jwt_expiration="{{ .Integration.MQTT.Auth.GCPCloudIoTCore.JWTExpiration }}"

    # JWT token key-file.
    #
    # Example command to generate a key-pair:
    #  $ ssh-keygen -t rsa -b 4096 -f private-key.pem
    #  $ openssl rsa -in private-key.pem -pubout -outform PEM -out public-key.pem
    #
    # Then point the setting below to the private-key.pem and associate the
    # public-key.pem with this device / gateway in Google Cloud IoT Core.
    jwt_key_file="{{ .Integration.MQTT.Auth.GCPCloudIoTCore.JWTKeyFile }}"


    # Azure IoT Hub
    #
    # This setting will preset uplink and downlink topics that will only
    # work with Azure IoT Hub service.
    [integration.mqtt.auth.azure_iot_hub]

    # Device connection string (symmetric key authentication).
    #
    # This connection string can be retrieved from the Azure IoT Hub device
    # details when using the symmetric key authentication type.
    device_connection_string="{{ .Integration.MQTT.Auth.AzureIoTHub.DeviceConnectionString }}"

    # Token expiration (symmetric key authentication).
    #
    # LoRa Gateway Bridge will generate a SAS token with the given expiration.
    # After the token has expired, it will generate a new one and trigger a
    # re-connect (only for symmetric key authentication).
    sas_token_expiration="{{ .Integration.MQTT.Auth.AzureIoTHub.SASTokenExpiration }}"

    # Device ID (X.509 authentication).
    #
    # This will be automatically set when a device connection string is given.
    # It must be set for X.509 authentication.
    device_id="{{ .Integration.MQTT.Auth.AzureIoTHub.DeviceID }}"

    # IoT Hub hostname (X.509 authentication).
    #
    # This will be automatically set when a device connection string is given.
    # It must be set for X.509 authentication.
    # Example: iot-hub-name.azure-devices.net
    hostname="{{ .Integration.MQTT.Auth.AzureIoTHub.Hostname }}"

    # Client certificates (X.509 authentication).
    #
    # Configure the tls_cert (certificate file) and tls_key (private-key file)
    # when the device is configured with X.509 authentication.
    tls_cert="{{ .Integration.MQTT.Auth.AzureIoTHub.TLSCert }}"
    tls_key="{{ .Integration.MQTT.Auth.AzureIoTHub.TLSKey }}"


# Metrics configuration.
[metrics]

  # Metrics stored in Prometheus.
  #
  # These metrics expose information about the state of the LoRa Gateway Bridge
  # instance like number of messages processed, number of function calls, etc.
  [metrics.prometheus]
  # Expose Prometheus metrics endpoint.
  endpoint_enabled={{ .Metrics.Prometheus.EndpointEnabled }}

  # The ip:port to bind the Prometheus metrics server to for serving the
  # metrics endpoint.
  bind="{{ .Metrics.Prometheus.Bind }}"


# Gateway meta-data.
#
# The meta-data will be added to every stats message sent by the LoRa Gateway
# Bridge.
[meta_data]

  # Static.
  #
  # Static key (string) / value (string) meta-data.
  [meta_data.static]
  # Example:
  # serial_number="A1B21234"
  {{ range $k, $v := .MetaData.Static }}
  {{ $k }}="{{ $v }}"
  {{ end }}


  # Dynamic meta-data.
  #
  # Dynamic meta-data is retrieved by executing external commands.
  # This makes it possible to for example execute an external command to
  # read the gateway temperature.
  [meta_data.dynamic]

  # Execution interval of the commands.
  execution_interval="{{ .MetaData.Dynamic.ExecutionInterval }}"

  # Max. execution duration.
  max_execution_duration="{{ .MetaData.Dynamic.MaxExecutionDuration }}"

  # Commands to execute.
  #
  # The value of the stdout will be used as the key value (string).
  # In case the command failed, it is ignored. In case the same key is defined
  # both as static and dynamic, the dynamic value has priority (as long as the)
  # command does not fail.
  [meta_data.dynamic.commands]
  # Example:
  # temperature="/opt/gateway-temperature/gateway-temperature.sh"
  {{ range $k, $v := .MetaData.Dynamic.Commands }}
  {{ $k }}="{{ $v }}"
  {{ end }}

# Executable commands.
#
# The configured commands can be triggered by sending a message to the
# LoRa Gateway Bridge.
[commands]
  # Example:
  # [commands.commands.reboot]
  # max_execution_duration="1s"
  # command="/usr/bin/reboot"
{{ range $k, $v := .Commands.Commands }}
  [commands.commands.{{ $k }}]
  max_execution_duration="{{ $v.MaxExecutionDuration }}"
  command="{{ $v.Command }}"
{{ end }}
`

var configCmd = &cobra.Command{
	Use:   "configfile",
	Short: "Print the LoRa Gateway Bridge configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		t := template.Must(template.New("config").Parse(configTemplate))
		err := t.Execute(os.Stdout, config.C)
		if err != nil {
			return errors.Wrap(err, "execute config template error")
		}
		return nil
	},
}
