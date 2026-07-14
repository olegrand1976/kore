DELETE FROM ai.ai_capabilities WHERE code IN (
    'tma.suggest_assignee',
    'tma.executive_summary',
    'cra.comment_summary',
    'budget.overrun_forecast',
    'conges.date_suggest',
    'publicsite.lead_scoring',
    'notifications.digest'
);
