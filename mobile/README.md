# Kore Mobile (Flutter)

Client mobile iOS / Android pour les parcours **CRA** et **congés**.  
Appelle l'API Go directement (`/api/v1/*`) avec `Authorization: Bearer`.

Références : `technical/foundation/14-flutter-mobile-client.md`, `technical/modules/16-mobile-flutter.md`.

## Prérequis

- Flutter 3.x (`flutter doctor`)
- Stack Kore locale : `make up` (API **8081**, frontend 3001)

## Configuration

| Variable | Description | Défaut dev |
| --- | --- | --- |
| `KORE_API_BASE_URL` | Base API (dart-define) | `http://10.0.2.2:8081/api/v1` (émulateur Android) |
| `KORE_TENANT_ID` | UUID tenant pour SSO | — |
| `OIDC_REDIRECT_URI` | Custom scheme | `kore://callback` |

Exemple iOS simulateur / appareil local :

```bash
flutter run \
  --dart-define=KORE_API_BASE_URL=http://localhost:8081/api/v1
```

## Installation

```bash
cd mobile
flutter pub get
```

## Lancer

```bash
flutter run
```

Compte seed : `ADM_admin` / `Admin123!`

## Tests & analyse

```bash
flutter analyze
flutter test
```

## Builds release

```bash
# Android (Play Store)
flutter build appbundle \
  --dart-define=KORE_API_BASE_URL=https://api.example.com/api/v1

# iOS (TestFlight)
flutter build ipa \
  --dart-define=KORE_API_BASE_URL=https://api.example.com/api/v1
```

## Structure

```
lib/
  main.dart, app.dart
  core/
    api/          # ApiClient (http + Bearer + refresh 401)
    auth/         # AuthRepository, OidcService (PKCE stub)
    theme/        # KoreColors / ThemeData (--kore-*)
    l10n/         # FR/EN map
  features/
    auth/         # Login password + SSO
    cra/          # Liste + éditeur semaine
    conges/       # Demandes, soldes, validation manager
    shared/       # KoreScaffold
```

## Endpoints consommés

| Module | Méthode | Chemin |
| --- | --- | --- |
| Auth | POST | `/auth/login`, `/auth/refresh`, `/auth/logout` |
| Auth OIDC | GET/POST | `/auth/oidc/authorize`, `/auth/oidc/callback` |
| CRA | GET | `/timesheets/recent`, `/timesheets?month=` |
| CRA | PUT/POST | `/timesheets/{id}/weeks/{week}`, `.../submit`, `.../validate` |
| Congés | GET/POST | `/leave-requests`, `/leave-balances` |
| Congés | POST | `/leave-requests/{id}/approve`, `.../reject` |

## OIDC PKCE

`OidcService.signInWithSso` est un **stub** : il construit l'URL via l'API Kore.  
Brancher `flutter_appauth` + intent-filter Android / URL scheme iOS pour le redirect `kore://callback`.

## iOS / Android native

Après `flutter create`, configurer :

- **Android** : `AndroidManifest.xml` intent-filter pour `kore://callback`
- **iOS** : `Info.plist` `CFBundleURLTypes` pour `kore`
