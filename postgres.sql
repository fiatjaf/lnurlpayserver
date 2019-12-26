CREATE TABLE backend (
  id text PRIMARY KEY, -- the node id
  kind text NOT NULL, -- spark, lnd, lntxbot etc.
  connection jsonb NOT NULL,

  CONSTRAINT connection_valid CHECK (
    char_length(connection::text) < 500
    AND (
      (kind = 'spark' AND (
        jsonb_typeof(connection->'endpoint') = 'string' AND
        jsonb_typeof(connection->'key') = 'string' AND
        jsonb_typeof(connection->'cert') = 'string'
      )) OR
      (kind = 'lnd' AND (
        jsonb_typeof(connection->'endpoint') = 'string' AND
        jsonb_typeof(connection->'macaroon') = 'string' AND
        jsonb_typeof(connection->'cert') = 'string'
      )) OR
      (kind = 'lntxbot' AND (
        jsonb_typeof(connection->'key') = 'string'
      ))
    )
  )
);

CREATE TABLE shop (
  id text PRIMARY KEY,
  backend text REFERENCES backend (id),
  key text NOT NULL DEFAULT md5(random()::text),
  message text,

  -- {"kind": "none"},
  -- {"kind": "sequential", "init": 0},
  -- {"kind": "hmac", "interval": 5, "key": "..."} (interval in minutes)
  verification jsonb NOT NULL DEFAULT '{"kind": "none"}',

  webhook text,
  telegram integer,

  CONSTRAINT verification_valid CHECK (
    char_length(verification::text) < 30
    AND (
      (verification->>'kind' = 'none') OR
      (verification->>'kind' = 'sequential' AND jsonb_typeof(verification->'init') = 'number') OR
      (verification->>'kind' = 'hmac' AND jsonb_typeof(verification->'interval') = 'number' AND jsonb_typeof(verification->'key') = 'string')
    )
  )
);

CREATE INDEX ON shop (key);

CREATE TABLE template (
  id text NOT NULL,
  shop text NOT NULL REFERENCES shop (id),
  path_params jsonb NOT NULL,
  query_params jsonb NOT NULL,
  description text NOT NULL, -- template
  image text, -- data-uri or nothing
  currency text NOT NULL DEFAULT 'sat', -- sat, usd, eur, brl etc.
  min_price text NOT NULL, -- formula
  max_price text NOT NULL, -- formula

  PRIMARY KEY (shop, id),
  CONSTRAINT arrays CHECK (
    jsonb_typeof(path_params->'array') AND
    jsonb_typeof(query_params->'array')
  ),
  CONSTRAINT currency_check CHECK (
    currency IN ('sat', 'eur', 'usd', 'gbp', 'cad', 'jpy')
  ),
  CONSTRAINT image_datauri CHECK (
    CASE WHEN image IS NOT NULL
      THEN (
        substring(image from '.*,') IN (
          'data:image/jpeg;base64,'
          'data:image/png;base64,'
        )
      )
      ELSE true
    END
  )
);

CREATE TABLE invoice (
  hash text PRIMARY KEY,
  preimage text UNIQUE NOT NULL,
  template text NOT NULL REFERENCES template (id),
  creation timestamp NOT NULL DEFAULT now(),
  payment timestamp, -- null when not paid
  amount_msat numeric(13) NOT NULL,
  bolt11 text NOT NULL,
);
