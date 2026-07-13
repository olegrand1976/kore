import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:kore_mobile/core/api/api_client.dart';
import 'package:kore_mobile/core/auth/auth_repository.dart';
import 'package:kore_mobile/core/auth/oidc_service.dart';
import 'package:kore_mobile/core/l10n/app_localizations.dart';
import 'package:kore_mobile/core/theme/kore_theme.dart';
import 'package:kore_mobile/features/auth/presentation/login_screen.dart';

void main() {
  testWidgets('LoginScreen smoke test — FR labels render', (tester) async {
    const baseUrl = 'http://localhost:8081/api/v1';
    final auth = AuthRepository(baseUrl: baseUrl);
    final api = ApiClient(baseUrl: baseUrl, authRepository: auth);
    final oidc = OidcService(apiClient: api, authRepository: auth);
    addTearDown(() {
      api.dispose();
      auth.dispose();
    });

    await tester.pumpWidget(
      MaterialApp(
        theme: KoreTheme.light(),
        localizationsDelegates: const [
          AppLocalizations.delegate,
          ...GlobalMaterialLocalizations.delegates,
        ],
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('fr'),
        home: LoginScreen(
          authRepository: auth,
          oidcService: oidc,
          onLoggedIn: () {},
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(tester.takeException(), isNull);
    final l10n = AppLocalizations(const Locale('fr'));
    expect(find.text(l10n.t('loginTitle')), findsOneWidget);
    expect(find.text(l10n.t('loginSubmit')), findsOneWidget);
    expect(find.text(l10n.t('loginSso')), findsOneWidget);
  });
}
