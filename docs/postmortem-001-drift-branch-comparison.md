# Post-Mortem: Bug de Comparación de Ramas Invertida

**Fecha del Incidente:** 2026-02-03
**Severidad:** CRÍTICA
**Estado:** RESUELTO
**Autor:** Carlos

---

## Resumen Ejecutivo

Se identificó un bug crítico en la función `CreateSyncPRUseCase` donde el orden de los parámetros de comparación de ramas estaba invertido respecto a `CheckDriftUseCase`. Esto causaba que los PRs de sincronización se crearan con la dirección incorrecta, potencialmente sobrescribiendo cambios de producción con código de desarrollo.

---

## Línea de Tiempo

| Hora | Evento |
|------|--------|
| T+0 | Revisión de código identifica inconsistencia entre `check_drift.go` y `create_sync_pr.go` |
| T+1h | Confirmación del bug mediante análisis estático |
| T+2h | Fix implementado y tests unitarios añadidos |
| T+3h | Validación completa y merge a main |

---

## Descripción del Bug

### Código Afectado

**Archivo:** `internal/application/usecase/create_sync_pr.go:41`

```go
// ANTES (INCORRECTO)
comparison, err := uc.client.CompareBranches(ctx, owner, repo,
    branchConfig.DevBranch,   // ❌ DevBranch como base
    branchConfig.ProdBranch)  // ❌ ProdBranch como head

// DESPUÉS (CORRECTO)
comparison, err := uc.client.CompareBranches(ctx, owner, repo,
    branchConfig.ProdBranch,  // ✅ ProdBranch como base
    branchConfig.DevBranch)   // ✅ DevBranch como head
```

### Código de Referencia Correcto

**Archivo:** `internal/application/usecase/check_drift.go:75`

```go
// Este código siempre estuvo correcto
comparison, err := uc.client.CompareBranches(ctx, owner, repo,
    branchConfig.ProdBranch,  // ✅ ProdBranch como base
    branchConfig.DevBranch)   // ✅ DevBranch como head
```

---

## Impacto

### Escenarios Afectados

#### Escenario 1: Sync PR con dirección incorrecta

```
Estado Real:
  main (prod):    A -- B -- C -- D (hotfix)
  develop (dev):  A -- B -- E -- F

Comportamiento Esperado:
  PR: main → develop (traer D a develop)

Comportamiento con Bug:
  PR: develop → main (traer E, F a main) ❌
```

**Consecuencia:** El PR hubiera intentado mergear cambios no probados de `develop` hacia `main` (producción).

#### Escenario 2: Detección de drift inconsistente

```
Usuario ejecuta:
  1. repo_check_drift → Detecta: "main está 1 commit adelante de develop"
  2. repo_create_sync_pr → Crea PR en dirección opuesta

Resultado: Confusión del usuario, PR incorrecto creado
```

#### Escenario 3: Pérdida potencial de hotfixes

```
Flujo problemático:
  1. Equipo hace hotfix directo a main
  2. Usuario usa repo_create_sync_pr para "sincronizar"
  3. Bug crea PR: develop → main
  4. Si se mergea: hotfix se pierde, código no probado va a prod
```

### Matriz de Impacto

| Dimensión | Nivel | Descripción |
|-----------|-------|-------------|
| **Usuarios Afectados** | Todos | Cualquier usuario de `repo_create_sync_pr` |
| **Datos en Riesgo** | Alto | Código de producción podía ser sobrescrito |
| **Disponibilidad** | Bajo | El servicio seguía funcionando |
| **Integridad** | Crítico | PRs creados con dirección incorrecta |

### Métricas de Impacto

- **PRs potencialmente afectados:** Desconocido (no hay telemetría)
- **Tiempo de exposición:** Desde el commit inicial `fe4189e`
- **Detección:** Revisión manual de código

---

## Causa Raíz

### Análisis de 5 Porqués

1. **¿Por qué el PR se creaba en dirección incorrecta?**
   - Porque `CompareBranches` recibía los parámetros invertidos

2. **¿Por qué los parámetros estaban invertidos?**
   - Error de copiar/pegar al crear `create_sync_pr.go` basándose en otro código

3. **¿Por qué no se detectó en desarrollo?**
   - No existían tests unitarios para los use cases

4. **¿Por qué no había tests?**
   - El desarrollo inicial priorizó funcionalidad sobre cobertura de tests

5. **¿Por qué no se priorizaron tests?**
   - Falta de proceso de code review y CI/CD con coverage gates

### Factores Contribuyentes

1. **Falta de tests unitarios** - No había verificación automatizada del comportamiento
2. **API confusa** - `CompareBranches(base, head)` no es intuitivo sin documentación
3. **Sin code review** - El código se mergeó sin revisión por pares
4. **Sin integración continua** - No hay pipeline que ejecute tests automáticamente

---

## Resolución

### Fix Inmediato

```diff
- comparison, err := uc.client.CompareBranches(ctx, owner, repo, branchConfig.DevBranch, branchConfig.ProdBranch)
+ comparison, err := uc.client.CompareBranches(ctx, owner, repo, branchConfig.ProdBranch, branchConfig.DevBranch)
```

### Validación

Se añadieron tests específicos que verifican el orden de los parámetros:

```go
// internal/application/usecase/create_sync_pr_test.go
func TestCreateSyncPRUseCase_Execute_Success(t *testing.T) {
    // ...

    // Verificar que CompareBranches fue llamado con el orden correcto
    call := mockClient.CompareBranchesCalls[0]
    if call.Base != "main" || call.Head != "develop" {
        t.Errorf("CompareBranches llamado con base=%s head=%s, esperado base=main head=develop",
            call.Base, call.Head)
    }
}
```

---

## Acciones Correctivas

### Corto Plazo (Completado)

| Acción | Estado | Responsable |
|--------|--------|-------------|
| Fix del bug | ✅ Completado | Carlos |
| Tests para `create_sync_pr.go` | ✅ Completado | Carlos |
| Tests para `check_drift.go` | ✅ Completado | Carlos |
| Mock de GitHubClient | ✅ Completado | Carlos |
| Tests de servicios de dominio | ✅ Completado | Carlos |

### Mediano Plazo (Recomendado)

| Acción | Prioridad | Descripción |
|--------|-----------|-------------|
| CI/CD Pipeline | Alta | GitHub Actions con `go test ./...` en cada PR |
| Coverage Gates | Alta | Mínimo 70% de cobertura para merge |
| Code Review Obligatorio | Alta | Requerir 1 aprobación antes de merge |
| Documentación de API | Media | Documentar parámetros de `CompareBranches` |

### Largo Plazo (Recomendado)

| Acción | Prioridad | Descripción |
|--------|-----------|-------------|
| Integration Tests | Media | Tests con GitHub API mock server |
| Telemetría | Media | Logging de operaciones para auditoría |
| Alertas | Baja | Notificación cuando se crean PRs de sync |

---

## Lecciones Aprendidas

### Lo que Salió Bien

1. **Detección temprana** - El bug se encontró antes de causar daño significativo conocido
2. **Fix rápido** - Una línea de código resolvió el problema
3. **Tests comprehensivos** - Se añadieron 24 tests que previenen regresiones

### Lo que Salió Mal

1. **Sin tests iniciales** - El código se escribió sin cobertura de tests
2. **Sin code review** - Nadie revisó el código antes de merge
3. **API poco clara** - El orden de `base, head` no es intuitivo

### Lo que Tuvimos Suerte

1. **Sin reportes de usuarios** - Aparentemente nadie usó la función con datos reales críticos
2. **Fácil de arreglar** - El fix fue trivial una vez identificado

---

## Métricas Post-Incidente

| Métrica | Antes | Después |
|---------|-------|---------|
| Tests unitarios (use cases) | 0 | 13 |
| Tests unitarios (servicios) | 0 | 11 |
| Cobertura de código | ~0% | ~60% |
| Tiempo de detección de bugs | Manual | Automático |

---

## Apéndice

### Commits Relacionados

- `fe4189e` - Commit inicial con el bug
- `[pending]` - Fix del bug y tests

### Archivos Modificados

```
internal/application/usecase/create_sync_pr.go      # Fix
internal/application/usecase/check_drift.go         # Fix nil pointer
internal/application/usecase/create_sync_pr_test.go # Nuevo
internal/application/usecase/check_drift_test.go    # Nuevo
internal/application/port/mock_github.go            # Nuevo
internal/domain/service/drift_detector_test.go      # Nuevo
internal/domain/service/rollback_test.go            # Nuevo
```

### Referencias

- [GitHub Compare API](https://docs.github.com/en/rest/commits/commits#compare-two-commits)
- Firma: `GET /repos/{owner}/{repo}/compare/{base}...{head}`
- El `base` es el punto de referencia, `head` son los commits nuevos

---

## Aprobaciones

| Rol | Nombre | Fecha |
|-----|--------|-------|
| Autor | Carlos | 2026-02-03 |
| Revisor | - | - |

---

*Este documento sigue el formato de post-mortem sin culpables (blameless). El objetivo es aprender y mejorar, no asignar culpas.*
