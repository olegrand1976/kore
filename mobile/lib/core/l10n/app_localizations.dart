import 'package:flutter/material.dart';

typedef L10nLookup = String Function(String key);

class AppLocalizations {
  AppLocalizations(this.locale);

  final Locale locale;

  static const supportedLocales = [Locale('fr'), Locale('en')];

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  static final Map<String, Map<String, String>> _strings = {
    'fr': {
      'appTitle': 'Kore',
      'loginTitle': 'Connexion',
      'loginPassword': 'Identifiant',
      'loginPasswordHint': 'Mot de passe',
      'loginSubmit': 'Se connecter',
      'loginSso': 'Connexion SSO',
      'loginTenant': 'Tenant ID (UUID)',
      'loginTenantRequired': 'Tenant ID requis pour le SSO',
      'loginError': 'Identifiants invalides',
      'navCra': 'CRA',
      'navConges': 'Congés',
      'craTitle': 'Mes CRA',
      'craEmpty': 'Aucun CRA récent',
      'craWeek': 'Semaine',
      'craStatus': 'Statut',
      'craSubmit': 'Soumettre la semaine',
      'craSave': 'Enregistrer',
      'congesTitle': 'Mes demandes',
      'congesNew': 'Nouvelle demande',
      'congesType': 'Type de congé',
      'congesFrom': 'Du',
      'congesTo': 'Au',
      'congesSubmit': 'Envoyer la demande',
      'congesEmpty': 'Aucune demande',
      'congesBalances': 'Soldes',
      'congesValidation': 'Validation',
      'congesApprove': 'Approuver',
      'congesReject': 'Refuser',
      'congesPending': 'En attente',
      'loading': 'Chargement…',
      'errorGeneric': 'Une erreur est survenue',
      'logout': 'Déconnexion',
    },
    'en': {
      'appTitle': 'Kore',
      'loginTitle': 'Sign in',
      'loginPassword': 'Login',
      'loginPasswordHint': 'Password',
      'loginSubmit': 'Sign in',
      'loginSso': 'SSO sign in',
      'loginTenant': 'Tenant ID (UUID)',
      'loginTenantRequired': 'Tenant ID required for SSO',
      'loginError': 'Invalid credentials',
      'navCra': 'Timesheets',
      'navConges': 'Leave',
      'craTitle': 'My timesheets',
      'craEmpty': 'No recent timesheets',
      'craWeek': 'Week',
      'craStatus': 'Status',
      'craSubmit': 'Submit week',
      'craSave': 'Save',
      'congesTitle': 'My requests',
      'congesNew': 'New request',
      'congesType': 'Leave type',
      'congesFrom': 'From',
      'congesTo': 'To',
      'congesSubmit': 'Submit request',
      'congesEmpty': 'No requests',
      'congesBalances': 'Balances',
      'congesValidation': 'Validation',
      'congesApprove': 'Approve',
      'congesReject': 'Reject',
      'congesPending': 'Pending',
      'loading': 'Loading…',
      'errorGeneric': 'Something went wrong',
      'logout': 'Sign out',
    },
  };

  String t(String key) {
    final lang = locale.languageCode;
    return _strings[lang]?[key] ?? _strings['fr']![key] ?? key;
  }
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  bool isSupported(Locale locale) =>
      AppLocalizations.supportedLocales
          .any((l) => l.languageCode == locale.languageCode);

  @override
  Future<AppLocalizations> load(Locale locale) async {
    return AppLocalizations(locale);
  }

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}
