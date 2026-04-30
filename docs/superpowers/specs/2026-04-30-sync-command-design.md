# `ggg sync` — Design

## Scopo

Comando unico che porta tutti i repository configurati allo stato target in un solo passaggio: clona i mancanti e fa pull di quelli puliti. Sostituisce il flusso manuale `ggg clone && ggg pull`.

## Comportamento utente

```
$ ggg sync [filter]
Sync 5 repositories? (2 to clone, 3 to pull) [Y/n]

⠋ Cloning 2 repositories...

  ✓ git@github.com:foo/bar.git → /Users/n/Developer/foo/bar
  ✓ git@github.com:foo/baz.git → /Users/n/Developer/foo/baz

⠋ Pulling 3 repositories...

  ✓ git@github.com:foo/qux.git
  ✓ git@github.com:foo/quux.git
  ✗ git@github.com:foo/corge.git  network error

  ↳ skipped: foo/dirty-repo (dirty)
```

## Argomenti e flag

Stesso contratto di `clone` e `pull`:

- Posizionale `[filter]` (opzionale)
- `--filter/-f <string>`
- `--group/-g <name>`

Nessun flag nuovo.

## Categorizzazione dei repo

A partire dalla lista risolta da `resolveBulkRepos`, ogni repo viene assegnato a una di queste categorie (in ordine di valutazione):

1. **Path error:** `repo.FullPath` fallisce → skip silenzioso (consistente con `clone`/`pull` attuali).
2. **Da clonare:** non clonato (`!repo.IsCloned`).
3. **Da pullare:** clonato e clean (`!repo.IsDirty`).
4. **Skipped — dirty:** clonato ma dirty.

I repo in (1) sono trasparenti all'utente. Quelli in (4) compaiono nel report finale.

## Conferma unica

Una sola call a `confirmBulkAction` con messaggio dinamico in base ai conteggi:

- Entrambi > 0: `"Sync %d repositories? (%d to clone, %d to pull)"`
- Solo clone:   `"Sync %d repositories? (%d to clone)"`
- Solo pull:    `"Sync %d repositories? (%d to pull)"`

Il totale `%d` è `cloneCount + pullCount`. I repo dirty non sono nel totale (non sono "azioni" da confermare).

`confirmBulkAction` rispetta il contratto esistente: la conferma scatta solo se non c'è filtro/positional, in linea con gli altri comandi multi-repo.

## Esecuzione (fasi seriali)

1. **Se** `len(cloneJobs) > 0`: `runParallelWithSpinner` con titolo
   - `"Cloning N repositories..."` se N>1
   - `"Cloning <url>..."` se N==1

   Stampa risultati per repo (✓ / ✗) come fa `clone` oggi.

2. **Se** `len(pullJobs) > 0`: `runParallelWithSpinner` con titolo
   - `"Pulling M repositories..."` se M>1
   - `"Pulling <url>..."` se M==1

   Stampa risultati per repo come fa `pull` oggi.

3. **Se** `len(dirtySkipped) > 0`: stampa una sezione finale `↳ skipped: <repo-name> (dirty)`, una riga per repo.

Errori in fase 1 NON interrompono fase 2.

## Edge cases

| Caso | Comportamento |
|---|---|
| Tutto in sync (no clone, no pull, no dirty) | Stampa `"Everything is in sync."` ed esce con 0. Nessuna conferma. |
| Solo dirty (no clone, no pull) | Stampa la sezione skipped + messaggio informativo. Nessuna conferma. Esce con 0. |
| Filter non matcha nulla | `resolveBulkRepos` ritorna `len(repos)==0` → exit silenzioso (uguale a `clone`/`pull`). |
| Conferma rifiutata | Esce con 0 senza eseguire azioni. |

## Implementazione

- File nuovo: `cmd/sync.go`
- `GroupID: GroupRepo`
- Registrazione via `init()` come gli altri comandi.
- Riusa interamente: `resolveBulkRepos`, `confirmBulkAction`, `runParallelWithSpinner`, `repo.Clone`, `repo.Pull`, `cfg.ResolvePullStrategy`, `repo.FullPath`, `repo.IsCloned`, `repo.IsDirty`, gli stili `ui.*`.
- Nessuna modifica a `repo/`, `config/`, `cmd/helpers.go`, `cmd/clone.go`, `cmd/pull.go`.

### Schema del file

```go
package cmd

var syncCmd = &cobra.Command{
    Use:     "sync [filter]",
    Short:   "Clone missing repositories and pull clean ones",
    GroupID: GroupRepo,
    Args:    cobra.MaximumNArgs(1),
    RunE:    runSync,
}

func init() {
    syncCmd.Flags().StringP("group", "g", "", "Sync only repos in this group")
    syncCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
    rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
    // 1. resolveBulkRepos
    // 2. categorize → cloneJobs, pullJobs, dirtySkipped
    // 3. handle "everything in sync" / "only dirty" early-exit
    // 4. confirmBulkAction (combined message)
    // 5. phase 1: clone if cloneJobs > 0
    // 6. phase 2: pull if pullJobs > 0
    // 7. print dirtySkipped section
}
```

I tipi `cloneJob` / `pullJob` sono dichiarati locali a `runSync` (come fanno `clone.go` e `pull.go` per i loro tipi).

## Test

### Unit (`cmd/sync_test.go`)
- Categorizzazione dei repo: data una lista di repo + uno stato disco simulato (file system temporaneo con repo veri creati via `initTestRepo`-like helper o `t.TempDir()`), verifica che la funzione di categorizzazione produca i tre slice attesi. Table-driven.

### E2E (`tests/sync_test.go`)
- Scenario combinato: config con 3 repo — uno non clonato, uno clean, uno dirty. Verifica:
  - Il non clonato viene clonato.
  - Il clean riceve il pull.
  - Il dirty appare nella sezione skipped.
  - Exit code 0.
- Scenario "everything in sync": tutti clonati e clean → output `"Everything is in sync."`.
- Scenario filter: `ggg sync foo` agisce solo sui repo che matchano.

Riusa l'infrastruttura hermetica già esistente (`internal/testutil`, pattern dei test e2e di `clone`/`pull`).

## Documentazione

1. Aggiungere voce `sync` a `docs/commands.md` (gruppo Repo operations).
2. Aggiornare la tabella dei comandi in `README.md`.
3. Spuntare `Comando ggg sync` in `ROADMAP.md` v1.1.0.
4. Nessuna modifica a `AGENTS.md` (pattern già documentati).

## Out of scope

- Push (mai automatico in `sync`).
- Stash automatico dei dirty (l'utente può usare `ggg stash` esplicitamente).
- Flag tipo `--no-confirm` o `--clone-only` / `--pull-only` (esistono già `clone` e `pull` come comandi separati).
- Output JSON o machine-readable.
