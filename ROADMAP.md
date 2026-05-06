# Roadmap

## v1.0.0

- [x] Definire il formato della configurazione YAML (`~/.config/ggg.yaml`)
- [x] Integrare Cobra come framework CLI
- [x] Comando `ggg init` — genera un file di configurazione di esempio
- [x] Comando `ggg list` — mostra i repository configurati e il loro stato (clonato / non clonato)
- [x] Comando `ggg clone` — clona tutti i repository configurati
- [x] Comando `ggg clone <name>` — clona un singolo repository
- [x] Derivazione automatica del path di destinazione dall'URL del repo (es. `github.com/user/repo`)
- [x] Supporto `base_dir` nella configurazione per la directory radice dei clone
- [x] Gestione errori (repo già clonato, URL non valido, directory non scrivibile)
- [x] Aggiungere test unitari
- [x] Completare il README con istruzioni di installazione, utilizzo ed esempi
- [x] Comando `ggg pull` — pull su tutti i repo (solo se puliti)
- [x] Comando `ggg status` — mostra branch, dirty/clean, ahead/behind
- [x] Comando `ggg add <url>` — aggiunge un repo alla config da CLI
- [x] Comando `ggg remove <name>` — rimuove un repo dalla config
- [x] Clone e pull paralleli con spinner
- [x] Supporto gruppi/tag con flag `--group/-g`
- [x] Comando `ggg cd <name>` — shell integration (`eval $(ggg cd <name>)`)
- [x] Output stilizzato con Charm stack (lipgloss + huh)
- [ ] Build cross-platform (Linux, macOS, Windows)
- [ ] Prima release su GitHub con binari precompilati

## v1.1.0

- [x] Comando `ggg doctor` — health check: config valida, repo orfani, remote raggiungibili
- [x] Comando `ggg outdated` — mostra i repo behind rispetto al remote
- [x] Comando `ggg open <name> [editor]` — apre il repo nell'editor (default: $EDITOR)
- [x] Comando `ggg browse <name>` — apre il repo nel browser
- [ ] Comando `ggg sync` — clona i mancanti + pull i puliti in un solo comando
- [x] Comando `ggg import [org]` — importa repo da GitHub via `gh` CLI
- [x] Comando `ggg export` — esporta la config in formato condivisibile
- [x] Shell alias — `gcd` function shell per navigare senza `eval`
- [ ] Completions dinamiche — autocompletamento nomi repo per bash/zsh/fish
- [ ] Config watch — flag `--watch` su status con refresh periodico
- [ ] Notifiche dirty — integrazione con prompt shell (starship/p10k)

## v1.2.0

- [x] Comando `ggg add <url> --clone` — aggiunge e clona in un colpo solo
- [x] Comando `ggg add <url> --group <name>` — specificare gruppo e path da CLI
- [x] Comando `ggg list --groups` — mostra i gruppi disponibili
- [ ] Comando `ggg rename <old> <new>` — rinomina il path/alias di un repo
- [x] Comando `ggg stash [name]` — esegue `git stash` su tutti i repo dirty
- [ ] Comando `ggg branch [name]` — mostra o filtra i repo per branch corrente
- [x] Comando `ggg checkout <branch> [name]` — checkout di un branch su tutti i repo
- [x] Comando `ggg validate` — validazione approfondita della config (URL duplicati, path conflittuali)
- [x] Comando `ggg diff [name]` — riassunto dei file modificati in tutti i repo dirty
- [ ] Supporto multi-config — merge di più file YAML (es. `work.yaml` + `personal.yaml`)
