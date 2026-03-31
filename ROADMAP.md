# Roadmap

## v1.0.0

- [ ] Definire il formato della configurazione YAML (`~/.config/ggg.yaml`)
- [ ] Integrare Cobra come framework CLI
- [ ] Comando `ggg init` — genera un file di configurazione di esempio
- [ ] Comando `ggg list` — mostra i repository configurati e il loro stato (clonato / non clonato)
- [ ] Comando `ggg clone` — clona tutti i repository configurati
- [ ] Comando `ggg clone <name>` — clona un singolo repository
- [ ] Derivazione automatica del path di destinazione dall'URL del repo (es. `github.com/user/repo`)
- [ ] Supporto `base_dir` nella configurazione per la directory radice dei clone
- [ ] Gestione errori (repo già clonato, URL non valido, directory non scrivibile)
- [ ] Aggiungere test unitari
- [ ] Completare il README con istruzioni di installazione, utilizzo ed esempi
- [ ] Build cross-platform (Linux, macOS, Windows)
- [ ] Prima release su GitHub con binari precompilati
