INSERT INTO ai.ai_capabilities (code, risk_class, annex_iii, wave) VALUES
    ('tma.suggest_assignee', 'limited', FALSE, 2),
    ('tma.executive_summary', 'limited', FALSE, 2),
    ('cra.comment_summary', 'limited', FALSE, 2),
    ('budget.overrun_forecast', 'minimal', FALSE, 2),
    ('conges.date_suggest', 'limited', FALSE, 3),
    ('publicsite.lead_scoring', 'minimal', FALSE, 3),
    ('notifications.digest', 'limited', FALSE, 3),
    ('mobile.voice_cra', 'limited', FALSE, 4)
ON CONFLICT (code) DO NOTHING;
