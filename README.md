# mqtt-lamarzocco

Control your La Marzocco espresso machine's brew-by-weight settings via MQTT and web interface.

## Features

- Switch between Dose 1, Dose 2, and Continuous modes
- MQTT integration for home automation
- Web interface for manual control
- Real-time status updates via Server-Sent Events

## Supported Machines

This application works with WiFi-enabled La Marzocco machines that support brew-by-weight:
- Linea Mini
- Linea Micra

## Prerequisites

- La Marzocco machine registered in the official La Marzocco Home app
- MQTT broker (e.g., Mosquitto)
- Docker (recommended) or Go 1.24+

## Configuration

Create a configuration file at `production/config/config.json`:

```json
{
  "mqtt": {
    "url": "tcp://your-mqtt-broker:1883",
    "topic": "home/lamarzocco",
    "qos": 2,
    "retain": true
  },
  "lamarzocco": {
    "username": "your-email@example.com",
    "password": "your-password",
    "polling_interval": 30
  },
  "web": {
    "enabled": true,
    "port": 8080
  },
  "loglevel": "info"
}
```

### Configuration Options

| Option | Description |
|--------|-------------|
| `mqtt.url` | MQTT broker URL |
| `mqtt.topic` | Base topic for MQTT messages |
| `mqtt.qos` | MQTT Quality of Service (0, 1, or 2) |
| `mqtt.retain` | Retain MQTT messages |
| `lamarzocco.username` | Your La Marzocco account email |
| `lamarzocco.password` | Your La Marzocco account password |
| `lamarzocco.polling_interval` | Status polling interval in seconds |
| `web.enabled` | Enable/disable web interface |
| `web.port` | Web server port |
| `loglevel` | Log level (debug, info, warn, error) |

### Environment Variable Substitution

You can use environment variables in the config file:

```json
{
  "lamarzocco": {
    "username": "${LM_USERNAME}",
    "password": "${LM_PASSWORD}"
  }
}
```

## MQTT Interface

### Topics

| Topic | Direction | Description |
|-------|-----------|-------------|
| `home/lamarzocco/status` | Publish | Current machine status |
| `home/lamarzocco/set` | Subscribe | Commands to set mode |

### Status Message

```json
{
  "mode": "Dose1",
  "connected": true,
  "serial": "MI012345",
  "model": "LINEA MINI 2023"
}
```

### Command Message

```json
{"mode": "Dose1"}
```

Valid modes: `Dose1`, `Dose2`, `Continuous`

## Web Interface

Access the web interface at `http://localhost:8080`

Features:
- Current mode display
- Mode selection buttons
- Real-time updates
- Dark/light theme toggle

## Running with Docker

```bash
docker run -d \
  -v /path/to/config.json:/var/lib/mqtt-lamarzocco/config.json \
  -p 8080:8080 \
  pharndt/mqtt-lamarzocco:latest
```

Or with docker-compose:

```yaml
version: '3.8'
services:
  mqtt-lamarzocco:
    image: pharndt/mqtt-lamarzocco:latest
    volumes:
      - ./config.json:/var/lib/mqtt-lamarzocco/config.json
    ports:
      - "8080:8080"
    restart: unless-stopped
```

## Building from Source

### Prerequisites

- Go 1.24+
- Node.js 22+
- pnpm

### Build

```bash
cd app
make build
```

### Run

```bash
./mqtt-lamarzocco /path/to/config.json
```

## Home Assistant Integration

### MQTT Sensor

```yaml
mqtt:
  sensor:
    - name: "La Marzocco Dose Mode"
      state_topic: "home/lamarzocco/status"
      value_template: "{{ value_json.mode }}"
      json_attributes_topic: "home/lamarzocco/status"
```

### MQTT Select

```yaml
mqtt:
  select:
    - name: "La Marzocco Dose Mode"
      command_topic: "home/lamarzocco/set"
      command_template: '{"mode": "{{ value }}"}'
      options:
        - "Dose1"
        - "Dose2"
        - "Continuous"
      state_topic: "home/lamarzocco/status"
      value_template: "{{ value_json.mode }}"
```

## API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/status` | GET | Get current status |
| `/api/mode` | POST | Set dose mode |
| `/api/events` | GET | SSE stream |

## License

MIT

## Acknowledgments

- [pylamarzocco](https://github.com/zweckj/pylamarzocco) - Python library that helped understand the API
