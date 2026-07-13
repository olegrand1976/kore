# 05 — Patterns UI IA

> Transparence Art. 50, supervision Art. 14.
> Charte : [`documentation/CHARTE_GRAPHIQUE.md`](../../documentation/CHARTE_GRAPHIQUE.md).

## 1. Composants

### `AppAiBadge`

Badge discret indiquant contenu ou interaction IA.

```vue
<AppAiBadge variant="generated" />
<AppAiBadge variant="assistant" />
```

Variants :

| Variant | i18n | Usage |
| --- | --- | --- |
| `generated` | `ai.badge.generated` | Texte généré par IA |
| `assistant` | `ai.badge.assistant` | Interaction chatbot |

Styles : tokens `--kore-*`, cohérent `AppBadge`.

### `AppAiSuggestion`

Pattern accept/modify/reject pour suggestions injectées dans formulaires.

```vue
<AppAiSuggestion
  :loading="generating"
  @accept="applyDraft"
  @reject="dismissDraft"
>
  <AppAiBadge variant="generated" />
  <!-- preview content -->
</AppAiSuggestion>
```

## 2. Règles UX

1. **Toute sortie IA modifiable** avant enregistrement métier
2. **Bouton explicite** « Générer brouillon » — pas de génération auto au chargement
3. **Disclaimer** sous preview : `ai.disclaimer`
4. **Mobile ≤768px** : actions empilées pleine largeur
5. **Disabled** si tenant IA non activé — message `ai.disabled_tenant`

## 3. i18n (fr.json / en.json)

```json
{
  "ai": {
    "badge": {
      "generated": "Généré par IA",
      "assistant": "Assistant IA"
    },
    "disclaimer": "Vérifiez le contenu avant enregistrement. L'IA peut se tromper.",
    "generate_draft": "Générer brouillon",
    "generating": "Génération…",
    "accept": "Appliquer",
    "reject": "Ignorer",
    "disabled_tenant": "L'assistance IA n'est pas activée pour votre organisation.",
    "notice": { ... }
  }
}
```

## 4. Intégrations par écran

| Écran | Composants |
| --- | --- |
| `tma/[id].vue` | `AnalysisEditor` + badge + generate |
| `dashboard/index.vue` | briefing card + badge |
| `conges/validation.vue` | contexte manager collapsible |
| `budget/[id].vue` | estimation suggest |
| Landing / chat | badge assistant + disclosure Art. 50 |

## 5. Accessibilité

- `role="status"` sur loading génération
- `aria-live="polite"` pour briefing dashboard
- Contraste badge conforme charte

## 6. Definition of Done

- [ ] `AppAiBadge.vue` créé
- [ ] `AppAiSuggestion.vue` créé (optionnel V1 — pattern inline acceptable)
- [ ] Clés i18n fr/en
- [ ] Test responsive 320px / 768px
