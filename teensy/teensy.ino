#include <Felica.h>

#define BUFFER_SIZE 64
#define MAX_BLOCKS 64

#define CMD_POLL 0x02
#define CMD_SYSTEMS 0x04
#define CMD_SERVICES 0x06
#define CMD_BLOCKS 0x08

#define STATUS_OK 0x00
#define STATUS_CONTINUE 0x01
#define STATUS_ERR 0x02

PN532 pn532(&Wire);
Felica felica(&pn532);

void setup() {
  while (!felica.begin()) {
    delay(1000);
  }
}

uint8_t buffer[BUFFER_SIZE];
uint8_t data[MAX_BLOCKS][FELICA_BLOCK_SIZE];

void loop() {
  if (RawHID.recv(buffer, 0) > 0) {
    switch (buffer[0]) {
    case CMD_POLL:
      poll();
      break;
    case CMD_SYSTEMS:
      systems();
      break;
    case CMD_SERVICES:
      services(buffer[1]);
      break;
    case CMD_BLOCKS:
      blocks(buffer[1]);
      break;
    }
  }
}

// [0]      CMD_POLL
// [1]      STATUS
// [2]      FOUND
// [3...10] CARD ID
void poll() {
  buffer[0] = CMD_POLL + 1;
  buffer[1] = STATUS_OK;
  if (felica.cmd_Polling()) {
    buffer[2] = 1;
    memcpy(buffer + 3, felica.IDm, FELICA_ID_LENGTH);
  } else {
    buffer[2] = 0;
  }
  RawHID.send(buffer, 100);
}

// [0]        CMD_SYSTEMS
// [1]        STATUS
// [2]        N
// [N+3..N+4] N-SYSTEM ID
void systems() {
  uint8_t system;
  buffer[0] = CMD_SYSTEMS + 1;
  if (felica.cmd_RequestSystemCode()) {
    buffer[1] = STATUS_OK;
    buffer[2] = felica.system_count;
    for (system = 0; system < felica.system_count; ++system) {
      buffer[2 * system + 3] = felica.systems[system] >> 8;
      buffer[2 * system + 4] = felica.systems[system];
    }
  } else {
    buffer[1] = STATUS_ERR;
  }
  RawHID.send(buffer, 100);
}

// [0]        CMD_SERVICES
// [1]        STATUS
// [2]        N
// [N+3..N+4] N-SERVICE ID
void services(uint8_t system) {
  uint8_t idx, service;
  bool msb;
  buffer[0] = CMD_SERVICES + 1;
  if (felica.cmd_SearchServiceCode(system)) {
    buffer[1] = STATUS_OK;
    buffer[2] = felica.service_count;
    for (idx = 3, service = 0; service < felica.service_count; ++service) {
      msb = true;
      do {
        if (BUFFER_SIZE == idx) {
          buffer[1] = STATUS_CONTINUE;
          RawHID.send(buffer, 100);
          buffer[1] = STATUS_OK;
          idx = 2;
        }
        if (msb) {
          buffer[idx] = felica.services[service] >> 8;
        } else {
          buffer[idx] = felica.services[service];
        }
        ++idx;
        msb = !msb;
      } while (!msb);
    }
  } else {
    buffer[1] = STATUS_ERR;
  }
  RawHID.send(buffer, 100);
}

// [0]         CMD_BLOCKS
// [1]         STATUS
// [2]         N
// [N+3..N+18] N-BLOCK DATA
void blocks(uint8_t service) {
  uint8_t cnt, idx, block, b;
  buffer[0] = CMD_BLOCKS + 1;
  if (felica.can_read_service(service)) {
    cnt = MAX_BLOCKS;
    if (felica.read_service(service, &cnt, data)) {
      buffer[1] = STATUS_OK;
      buffer[2] = cnt;
      for (idx = 3, block = 0; block < cnt; ++block) {
        for (b = 0; b < FELICA_BLOCK_SIZE; ++idx, ++b) {
          if (BUFFER_SIZE == idx) {
            buffer[1] = STATUS_CONTINUE;
            RawHID.send(buffer, 100);
            buffer[1] = STATUS_OK;
            idx = 2;
          }
          buffer[idx] = data[block][b];
        }
      }
    } else {
      buffer[1] = STATUS_ERR;
    }
  } else {
    buffer[1] = STATUS_ERR + 1;
  }
  RawHID.send(buffer, 100);
}
