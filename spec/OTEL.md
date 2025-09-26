# AIS OpenTelemetry Conventions

## Spans
- uia.create (optional)
- apa.plan
- apr.verify
- ibe.guard
- tool.invoke

## Attributes
- ais.uia.id, ais.apa.id, ais.apr.id, ais.ibe.id
- ais.apa.step.id, ais.tool.name
- ais.apr.method, ais.apr.coverage, ais.apr.risk
- ais.guard.result (ok|deny)

## Links
- Link `apa.plan` → `apr.verify` → `ibe.guard` → `tool.invoke` with `spanLinks` referencing ids.

## Events
- guard.check: { checks: [alignment, risk, tca, data] }
- audit.emit: { ok: bool }
