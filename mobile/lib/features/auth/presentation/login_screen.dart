import 'package:flutter/material.dart';

import '../../../core/auth/auth_repository.dart';
import '../../../core/auth/oidc_service.dart';
import '../../../core/l10n/app_localizations.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({
    super.key,
    required this.authRepository,
    required this.oidcService,
    required this.onLoggedIn,
  });

  final AuthRepository authRepository;
  final OidcService oidcService;
  final VoidCallback onLoggedIn;

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _loginCtrl = TextEditingController(text: 'ADM_admin');
  final _passwordCtrl = TextEditingController(text: 'Admin123!');
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _loginCtrl.dispose();
    _passwordCtrl.dispose();
    super.dispose();
  }

  Future<void> _submitPassword() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await widget.authRepository.loginWithPassword(
        login: _loginCtrl.text.trim(),
        password: _passwordCtrl.text,
      );
      widget.onLoggedIn();
    } catch (_) {
      setState(() => _error = AppLocalizations.of(context).t('loginError'));
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _submitSso() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await widget.oidcService.signInWithSso(
        tenantId: const String.fromEnvironment(
          'KORE_TENANT_ID',
          defaultValue: '',
        ),
      );
      widget.onLoggedIn();
    } catch (e) {
      setState(() => _error = e.toString());
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 420),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    l10n.t('loginTitle'),
                    style: Theme.of(context).textTheme.headlineMedium,
                  ),
                  const SizedBox(height: 24),
                  TextField(
                    controller: _loginCtrl,
                    decoration: InputDecoration(labelText: l10n.t('loginPassword')),
                    textInputAction: TextInputAction.next,
                    autofillHints: const [AutofillHints.username],
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: _passwordCtrl,
                    decoration:
                        InputDecoration(labelText: l10n.t('loginPasswordHint')),
                    obscureText: true,
                    onSubmitted: (_) => _submitPassword(),
                    autofillHints: const [AutofillHints.password],
                  ),
                  if (_error != null) ...[
                    const SizedBox(height: 12),
                    Text(
                      _error!,
                      style: TextStyle(color: Theme.of(context).colorScheme.error),
                    ),
                  ],
                  const SizedBox(height: 24),
                  ElevatedButton(
                    onPressed: _loading ? null : _submitPassword,
                    child: _loading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Text(l10n.t('loginSubmit')),
                  ),
                  const SizedBox(height: 12),
                  OutlinedButton(
                    onPressed: _loading ? null : _submitSso,
                    child: Text(l10n.t('loginSso')),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
