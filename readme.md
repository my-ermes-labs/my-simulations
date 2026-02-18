# my-simulations

Simulazioni sperimentali per il framework Ermes. Ogni test misura un aspetto diverso del ciclo di vita delle sessioni distribuite: **migrazione**, **offloading** e **onloading**.

---

## Indice

- [Requisiti comuni](#requisiti-comuni)
- [Topologia dell'infrastruttura](#topologia-dellinfrastruttura)
- [Setup delle VM](#setup-delle-vm)
- [Simulazione delle latenze di rete](#simulazione-delle-latenze-di-rete)
- [Test: Migration](#test-migration)
- [Test: Offloading](#test-offloading)
- [Test: Onloading](#test-onloading)

---

## Requisiti comuni

### Sulla macchina locale (da cui si lancia tutto)

- Go >= 1.22
- `faas-cli` installato e nel PATH
- `ansible` >= 2.12
- Accesso SSH senza password verso tutte le VM (chiave pubblica copiata)
- Docker (per il build delle immagini FaaS)

### Su ogni VM

- Linux (testato su Ubuntu 22.04 / Debian 11, ARM o x86)
- `runc` installato
- `faasd` installato e in esecuzione sulla porta `8080`
- Redis in esecuzione sulla porta `6379` (immagine `ariannadragoniii/arm-redis:latest`)
- Funzioni FaaS deployate tramite `my-ermes-toolkit`

---

## Topologia dell'infrastruttura

I test usano 3 nodi organizzati gerarchicamente:

```
Italy (cloud, parent)         <ip-1>
├── Ravenna (edge 1)          <ip-1.1>
└── Milan (edge 2)            <ip-1.2>
```

I placeholder `<ip-1>`, `<ip-1.1>`, `<ip-1.2>` vanno sostituiti con i tuoi IP reali ovunque compaiano (vedi sezione di ogni test).

---

## Setup delle VM

### 1. Configura l'inventario Ansible

Modifica `../my-ermes-toolkit/inventory.ini`:

```ini
[my_hosts]
<ip-1>   ansible_user=italy   ansible_become=yes
<ip-1.1> ansible_user=ravenna ansible_become=yes
<ip-1.2> ansible_user=milan   ansible_become=yes
```

### 2. Configura l'infrastruttura

Modifica `infrastructure.json` nella root di `my-simulations/`:

```json
{
  "areaName": "Italy",        "host": "<ip-1>",
  "areaName": "Ravenna (PC)", "host": "<ip-1.1>",
  "areaName": "Milan",        "host": "<ip-1.2>"
}
```

### 3. Deploy delle funzioni FaaS su ogni nodo

Per ogni nodo, passa il JSON del nodo come variabile Ansible. Dal progetto `my-ermes-toolkit`:

```bash
cd ../my-ermes-toolkit

# Deploy su Ravenna
ansible-playbook -i inventory.ini deploy.yml \
  -e "target_hosts=<ip-1.1>" \
  -e "target_node=$(cat infrastructure.json | jq '.areas[0].areas[0]')"

# Deploy su Milan
ansible-playbook -i inventory.ini deploy.yml \
  -e "target_hosts=<ip-1.2>" \
  -e "target_node=$(cat infrastructure.json | jq '.areas[0].areas[1]')"

# Deploy su Italy
ansible-playbook -i inventory.ini deploy.yml \
  -e "target_hosts=<ip-1>" \
  -e "target_node=$(cat infrastructure.json | jq '.areas[0]')"
```

### 4. Verifica che le funzioni siano attive

```bash
faas-cli list --gateway http://<ip-1.1>:8080
faas-cli list --gateway http://<ip-1.2>:8080
faas-cli list --gateway http://<ip-1>:8080
```

Devono comparire: `hello-world`, `api`, `s-to-t`, `cdn-upload`, `cdn-download`, `migrate`.

---

## Simulazione delle latenze di rete

**Obbligatorio se le VM sono sulla stessa macchina fisica.** Senza latenza artificiale i risultati non sono rappresentativi.

Esegui su ogni VM (sostituisci `eth0` con il nome corretto dell'interfaccia, verificabile con `ip link`):

```bash
# Su Ravenna: 5ms verso Milan, 30ms verso Italy
sudo tc qdisc add dev eth0 root netem delay 5ms

# Su Milan: 5ms verso Ravenna, 30ms verso Italy  
sudo tc qdisc add dev eth0 root netem delay 5ms

# Su Italy: 30ms verso i nodi edge
sudo tc qdisc add dev eth0 root netem delay 30ms
```

Per simulare anche bandwidth limitata (più realistico per edge):
```bash
sudo tc qdisc add dev eth0 root netem delay 20ms rate 10mbit
```

Per rimuovere la latenza al termine degli esperimenti:
```bash
sudo tc qdisc del dev eth0 root
```

Valori di riferimento:

| Link | Latenza consigliata |
|---|---|
| Edge ↔ Edge (stessa area) | 2–5 ms |
| Edge → Cloud regionale | 20–40 ms |
| Edge → Cloud continentale | 80–120 ms |

---

## Test: Migration

**Cosa misura:** il tempo necessario a trasferire una sessione di dimensione variabile da Ravenna (edge) verso Italy (cloud). Esegue 200 migrazioni per ciascuna delle 8 dimensioni di sessione.

**Dimensioni testate:** 1, 256, 512, 1024, 2048, 3072, 4096, 5120 KB  
**Ripetizioni per dimensione:** 200  
**Output:** `migration/migration_resultss.csv` (`Session Size (KB)`, `Total Migration Time (ms)`)

### Prerequisiti specifici

- Redis su Ravenna deve contenere le chiavi di sessione pre-popolate (le crea `seed.go`)
- La funzione `migrate` deve essere deployata e attiva su Ravenna (`http://<ip-1.1>:8080/function/migrate`)

### IP da modificare

**File:** `migration/test.go`, riga con `ravennaNodeURL`:

```go
ravennaNodeURL = "http://<ip-1.1>:8080/function/migrate?size="
```

**File:** `migration/migration_trigger.go`, riga con `url.Parse`:

```go
url, err := url.Parse("http://localhost/ermes-api/migration")
// "localhost" è corretto — questo codice gira DENTRO la funzione FaaS su Ravenna.
// Cambia "node-area" con il nome del nodo parent in infrastructure.json:
query.Set("node-area", "Italy")
```

> `migration_trigger.go` è il codice che viene eseguito all'interno della funzione `migrate` su Ravenna. Va deployato come parte della funzione FaaS, non eseguito localmente.

### Come eseguire

```bash
cd migration

# Popola il Redis di Ravenna con le chiavi di sessione
# (eseguito direttamente su Ravenna o tramite SSH)
ERMES_NODE=$(cat ../infrastructure.json | jq -c '.areas[0].areas[0]') \
REDIS_HOST=<ip-1.1> \
go run seed.go

# Lancia il test dalla macchina locale
go run test.go
```

---

## Test: Offloading

**Cosa misura:** il response time percepito dai client durante un evento di offloading. 60 client fanno richieste continue per 20 secondi verso un edge node; durante questo periodo il nodo offloada automaticamente le sessioni eccessive verso il cloud. I response time vengono aggregati in finestre da 400ms.

**Client simultanei:** 60  
**Durata:** 20 secondi  
**Risorse chiamate:** `/test/speech-to-text`, `/test/cdn-upload`, `/test/cdn-download`  
**Output:** `offloading/slice_data.csv` (`Only200ResponseTime`, `Non200ResponseTime`)

### Prerequisiti specifici

- L'edge node configurato deve avere le soglie di offloading impostate in modo che l'offloading si attivi durante i 20 secondi di test (configura `resources.cpu` e `resources.io` con valori bassi in `infrastructure.json`)
- Le funzioni `speech-to-text`, `cdn-upload`, `cdn-download` devono essere attive sul nodo edge
- Un endpoint `test/create-session` deve essere esposto sul nodo

### IP da modificare

**File:** `offloading/test.go`, riga con `edgeNodeIP`:

```go
edgeNodeIP := "<ip-1.1>"   // IP del nodo edge (Ravenna o Milan)
```

### Come eseguire

```bash
cd offloading
go run test.go pretest.go
```

> `pretest.go` contiene codice commentato di test preliminari. Se vuoi usarlo, decommenta e configura gli URL.

---

## Test: Onloading

**Cosa misura:** il comportamento del client durante una sequenza di onloading. Invia 30 richieste GET sequenziali verso Milan; se il server risponde con redirect (HTTP 301/302), il client segue automaticamente il redirect verso il nuovo nodo. Registra per ogni richiesta: timestamp, codice di risposta, tempo totale (incluso eventuale redirect).

**Richieste:** 30 sequenziali  
**Output:** `onloading/responses.csv` (`Time`, `ResponseCode`, `ResponseTime`)

### Prerequisiti specifici

- Milan deve essere configurato per fare onloading della sessione verso Ravenna o Italy durante il test
- L'endpoint `/api/endpoint` deve essere esposto su Milan

### IP da modificare

**File:** `onloading/script.go`, riga con `url :=`:

```go
url := "http://<ip-1.2>/api/endpoint"   // IP di Milan
```

### Come eseguire

```bash
cd onloading
go run script.go
```

---

## Ordine consigliato degli esperimenti

1. Configura `infrastructure.json` e `inventory.ini` con i tuoi IP
2. Aggiungi le latenze di rete con `tc` su ogni VM
3. Deploya le funzioni FaaS con Ansible
4. Esegui **migration** (richiede seed del Redis)
5. Esegui **offloading** (verifica che il trigger di offload scatti durante i 20s)
6. Esegui **onloading**
7. Rimuovi le latenze con `tc qdisc del dev eth0 root`
8. Analizza i CSV prodotti

---

## Troubleshooting

| Problema | Soluzione |
|---|---|
| `migrate` non risponde | Verifica che faasd sia attivo: `systemctl status faasd` |
| Redis non raggiungibile | Controlla porta 6379: `redis-cli -h <ip> ping` |
| Offloading non si attiva | Abbassa le soglie `resources.cpu`/`resources.io` in `infrastructure.json` e rideploya |
| Redirect non seguiti | Verifica che il server risponda con header `Location` e codice 301/302 |
| Risultati troppo rapidi | Le latenze `tc` non sono attive o `eth0` è il nome sbagliato dell'interfaccia |
