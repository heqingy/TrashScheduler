package main

const yamlConfigTemplate = `
users:
- name: Zhang
  phone:
  - "+1**********"
  canpulled: 0
- name: Wang
  phone:
  - "+1**********"
  canpulled: 2
smsconfig:
  accid: ********************************
  token: ********************************
  number: "+1********"
fsmconfig:
  config:
    callinterval: 1h0m0s
    remindinterval: 1h0m0s
    notifysvc: null
  statepath: ./state.yml
`
