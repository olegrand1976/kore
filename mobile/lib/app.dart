import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:go_router/go_router.dart';

import 'core/api/api_client.dart';
import 'core/auth/auth_repository.dart';
import 'core/auth/oidc_service.dart';
import 'core/l10n/app_localizations.dart';
import 'core/theme/kore_theme.dart';
import 'features/auth/presentation/login_screen.dart';
import 'features/conges/data/leave_repository.dart';
import 'features/conges/presentation/leave_list_screen.dart';
import 'features/cra/data/cra_repository.dart';
import 'features/cra/presentation/cra_list_screen.dart';

/// Default API base — Android emulator loopback to host `make up` (port 8081).
const kDefaultApiBase = String.fromEnvironment(
  'KORE_API_BASE_URL',
  defaultValue: 'http://10.0.2.2:8081/api/v1',
);

class KoreAppScope extends InheritedWidget {
  const KoreAppScope({
    super.key,
    required this.authRepository,
    required this.apiClient,
    required this.oidcService,
    required this.craRepository,
    required this.leaveRepository,
    required super.child,
  });

  final AuthRepository authRepository;
  final ApiClient apiClient;
  final OidcService oidcService;
  final CraRepository craRepository;
  final LeaveRepository leaveRepository;

  static KoreAppScope of(BuildContext context) {
    return context.dependOnInheritedWidgetOfExactType<KoreAppScope>()!;
  }

  @override
  bool updateShouldNotify(KoreAppScope oldWidget) => false;
}

class KoreApp extends StatefulWidget {
  const KoreApp({super.key, this.apiBase = kDefaultApiBase});

  final String apiBase;

  @override
  State<KoreApp> createState() => _KoreAppState();
}

class _KoreAppState extends State<KoreApp> {
  late final AuthRepository _auth;
  late final ApiClient _api;
  late final OidcService _oidc;
  late final CraRepository _cra;
  late final LeaveRepository _leave;
  late final GoRouter _router;
  bool _canValidateLeave = false;

  @override
  void initState() {
    super.initState();
    _auth = AuthRepository(baseUrl: widget.apiBase);
    _api = ApiClient(baseUrl: widget.apiBase, authRepository: _auth);
    _oidc = OidcService(apiClient: _api, authRepository: _auth);
    _cra = CraRepository(_api);
    _leave = LeaveRepository(_api);
    _router = GoRouter(
      initialLocation: '/login',
      redirect: (context, state) async {
        final loggedIn = await _auth.isAuthenticated;
        final onLogin = state.matchedLocation == '/login';
        if (!loggedIn && !onLogin) return '/login';
        if (loggedIn && onLogin) return '/cra';
        return null;
      },
      routes: [
        GoRoute(
          path: '/login',
          builder: (context, state) => LoginScreen(
            authRepository: _auth,
            oidcService: _oidc,
            onLoggedIn: () async {
              final session = await _auth.loadSession();
              setState(() => _canValidateLeave = session?.canValidateLeave ?? false);
              _router.go('/cra');
            },
          ),
        ),
        GoRoute(
          path: '/cra',
          builder: (context, state) => CraListScreen(
            repository: _cra,
            onNavigateConges: () => _router.go('/conges'),
          ),
        ),
        GoRoute(
          path: '/conges',
          builder: (context, state) => LeaveListScreen(
            repository: _leave,
            onNavigateCra: () => _router.go('/cra'),
            canValidate: _canValidateLeave,
          ),
        ),
      ],
    );
    _auth.loadSession().then((session) {
      if (mounted) {
        setState(() => _canValidateLeave = session?.canValidateLeave ?? false);
      }
    });
  }

  @override
  void dispose() {
    _router.dispose();
    _api.dispose();
    _auth.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return KoreAppScope(
      authRepository: _auth,
      apiClient: _api,
      oidcService: _oidc,
      craRepository: _cra,
      leaveRepository: _leave,
      child: MaterialApp.router(
        title: 'Kore',
        theme: KoreTheme.light(),
        darkTheme: KoreTheme.dark(),
        themeMode: ThemeMode.system,
        localizationsDelegates: const [
          AppLocalizations.delegate,
          ...GlobalMaterialLocalizations.delegates,
        ],
        supportedLocales: AppLocalizations.supportedLocales,
        routerConfig: _router,
      ),
    );
  }
}
