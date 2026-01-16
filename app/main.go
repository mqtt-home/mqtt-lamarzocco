package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mqtt-home/mqtt-lamarzocco/config"
	"github.com/mqtt-home/mqtt-lamarzocco/lamarzocco"
	"github.com/mqtt-home/mqtt-lamarzocco/version"
	"github.com/mqtt-home/mqtt-lamarzocco/web"
	"github.com/philipparndt/go-logger"
	"github.com/philipparndt/mqtt-gateway/mqtt"
	"github.com/tidwall/gjson"
)

var client *lamarzocco.Client

func publishStatus(status lamarzocco.MachineStatus) {
	cfg := config.Get()
	topic := cfg.MQTT.Topic + "/status"

	data, err := json.Marshal(status)
	if err != nil {
		logger.Error("Failed to marshal status", err)
		return
	}

	mqtt.PublishAbsolute(topic, string(data), cfg.MQTT.Retain)
	logger.Debug("Published status", "topic", topic, "status", string(data))
}

func subscribeToCommands() {
	cfg := config.Get()
	topic := cfg.MQTT.Topic + "/set"

	logger.Info("Subscribing to MQTT commands", "topic", topic)

	mqtt.Subscribe(topic, func(topic string, payload []byte) {
		logger.Debug("Received MQTT command", "topic", topic, "payload", string(payload))

		cmd, err := lamarzocco.ParseCommand(payload)
		if err != nil {
			logger.Error("Failed to parse command", "error", err)
			return
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic in command processing", "panic", r)
				}
			}()

			// Handle dose1 command
			if cmd.HasDose1() {
				logger.Info("Setting dose1 weight", "weight", cmd.GetDose1())
				if err := client.SetDose("Dose1", cmd.GetDose1()); err != nil {
					logger.Error("Failed to set dose1", "error", err)
				}
			}

			// Handle dose2 command
			if cmd.HasDose2() {
				logger.Info("Setting dose2 weight", "weight", cmd.GetDose2())
				if err := client.SetDose("Dose2", cmd.GetDose2()); err != nil {
					logger.Error("Failed to set dose2", "error", err)
				}
			}

			// Handle mode command
			if cmd.HasMode() {
				mode := cmd.GetDoseMode()
				logger.Info("Setting dose mode", "mode", mode)
				if err := client.SetMode(mode); err != nil {
					logger.Error("Failed to set mode", "error", err)
				}
			}

			// Handle back flush command
			if cmd.HasBackFlush() {
				logger.Info("Starting back flush")
				if err := client.StartBackFlush(); err != nil {
					logger.Error("Failed to start back flush", "error", err)
				}
			}
		}()
	})
}

func matchValue(actual gjson.Result, expected interface{}) bool {
	if !actual.Exists() {
		return false
	}

	switch v := expected.(type) {
	case float64:
		return actual.Num == v
	case string:
		return actual.Str == v
	case bool:
		return actual.Bool() == v
	default:
		return actual.String() == v
	}
}

func subscribeToTriggers() {
	cfg := config.Get()

	if len(cfg.Triggers) == 0 {
		logger.Debug("No triggers configured")
		return
	}

	// Group triggers by topic
	triggersByTopic := make(map[string][]config.Trigger)
	for _, trigger := range cfg.Triggers {
		triggersByTopic[trigger.Topic] = append(triggersByTopic[trigger.Topic], trigger)
	}

	// Subscribe to each unique topic
	for topic, triggers := range triggersByTopic {
		subscribeTopic := topic    // capture topic for closure
		topicTriggers := triggers  // capture triggers for closure
		logger.Info("Subscribing to trigger topic", "topic", subscribeTopic, "triggers", len(topicTriggers))

		mqtt.Subscribe(subscribeTopic, func(msgTopic string, payload []byte) {
			logger.Info("Received trigger message", "topic", msgTopic, "payload_len", len(payload))

			payloadStr := string(payload)

			// Check each trigger for this topic
			for i, trigger := range topicTriggers {
				allMatch := true

				// Check all conditions
				for _, condition := range trigger.Conditions {
					result := gjson.Get(payloadStr, condition.Selector)
					logger.Debug("Checking condition",
						"selector", condition.Selector,
						"expected", condition.Value,
						"actual", result.Value(),
						"exists", result.Exists())
					if !matchValue(result, condition.Value) {
						allMatch = false
						break
					}
				}

				if allMatch {
					mode := lamarzocco.ParseDoseMode(trigger.Action.Mode)
					logger.Info("Trigger matched, setting dose mode",
						"trigger_index", i,
						"topic", msgTopic,
						"mode", mode)

					go func(m lamarzocco.DoseMode) {
						defer func() {
							if r := recover(); r != nil {
								logger.Error("Panic in trigger processing", "panic", r)
							}
						}()

						if err := client.SetMode(m); err != nil {
							logger.Error("Failed to set mode from trigger", "error", err)
						}
					}(mode)

					// Stop after first matching trigger
					return
				} else {
					logger.Debug("Trigger did not match", "trigger_index", i)
				}
			}

			logger.Debug("No trigger matched for message", "topic", msgTopic)
		})
	}

	logger.Info("Trigger subscriptions active", "topics", len(triggersByTopic), "triggers", len(cfg.Triggers))
}

func main() {
	logger.Info("mqtt-lamarzocco", version.Info())

	if len(os.Args) < 2 {
		logger.Error("No configuration file specified")
		os.Exit(1)
	}

	configFile := os.Args[1]
	logger.Info("Configuration file:", configFile)

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		logger.Error("Failed to load configuration", err)
		return
	}

	logger.SetLevel(cfg.LogLevel)

	// Start MQTT first (needed for status callback)
	mqtt.Start(cfg.MQTT, "lamarzocco_mqtt")

	// Initialize La Marzocco client
	client = lamarzocco.NewClient(
		cfg.LaMarzocco.Username,
		cfg.LaMarzocco.Password,
	)

	// Set callback to publish status on change
	client.SetStatusChangeCallback(publishStatus)

	// Connect to La Marzocco API
	logger.Info("Connecting to La Marzocco API...")
	if err := client.Connect(); err != nil {
		logger.Error("Failed to connect to La Marzocco API", err)
		return
	}

	// Publish initial status
	publishStatus(client.GetStatus())

	// Subscribe to commands
	subscribeToCommands()

	// Subscribe to configured triggers
	subscribeToTriggers()

	// Start polling for status updates
	stopPolling := make(chan struct{})
	go client.StartPolling(time.Duration(cfg.LaMarzocco.PollingInterval)*time.Second, stopPolling)

	// Start web server
	if !cfg.Web.Enabled {
		logger.Info("Web interface is disabled in the configuration")
	} else {
		logger.Info("Web interface enabled, starting web server")
		webServer := web.NewWebServer(client)
		go func() {
			err := webServer.Start(cfg.Web.Port)
			if err != nil {
				logger.Error("Failed to start web server", err)
			}
		}()
		logger.Info("Application is now ready. Web interface available at http://localhost:" + strconv.Itoa(cfg.Web.Port) + ". Press Ctrl+C to quit.")
	}

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

	close(stopPolling)
	logger.Info("Received quit signal")
}
