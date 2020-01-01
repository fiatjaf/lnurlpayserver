CREATE TABLE backend (
  id text PRIMARY KEY, -- the node id
  kind text NOT NULL, -- spark, lnd, lntxbot etc.
  connection jsonb NOT NULL,

  CONSTRAINT connection_length CHECK (char_length(connection::text) < 2000),
  CONSTRAINT connection_schema CHECK (
    (kind = 'spark' AND (
      jsonb_typeof(connection->'endpoint') = 'string' AND
      jsonb_typeof(connection->'key') = 'string' AND
      CASE WHEN connection ? 'cert'
        THEN jsonb_typeof(connection->'cert') = 'string'
        ELSE true
      END
    )) OR
    (kind = 'lnd' AND (
      jsonb_typeof(connection->'endpoint') = 'string' AND
      jsonb_typeof(connection->'macaroon') = 'string' AND
      CASE WHEN connection ? 'cert'
        THEN jsonb_typeof(connection->'cert') = 'string'
        ELSE true
      END
    )) OR
    (kind = 'lntxbot' AND (
      jsonb_typeof(connection->'key') = 'string'
    ))
  )
);

CREATE TABLE shop (
  id text PRIMARY KEY,
  backend text REFERENCES backend (id),
  key text NOT NULL DEFAULT md5(random()::text),
  message text,
  webhook text,

  -- {"kind": "none"}
  -- {"kind": "sequential", "init": 0, "words": ["pluc", "plec", "plic"]})
  -- {"kind": "hmac", "interval": 5, "key": "..."} (interval in minutes)
  verification jsonb NOT NULL DEFAULT '{"kind": "none"}',

  CONSTRAINT verification_length CHECK (char_length(verification::text) < 300),
  CONSTRAINT verification_schema CHECK (
    (verification->>'kind' = 'none') OR
    (verification->>'kind' = 'sequential' AND
      jsonb_typeof(verification->'init') = 'number' AND
      CASE WHEN verification ? 'words'
        THEN jsonb_typeof(verification->'words') = 'array'
        ELSE true
      END
    ) OR
    (verification->>'kind' = 'hmac' AND
      jsonb_typeof(verification->'interval') = 'number' AND
      jsonb_typeof(verification->'key') = 'string'
    )
  )
);

CREATE INDEX ON shop (key);

CREATE TABLE template (
  id text NOT NULL,
  shop text NOT NULL REFERENCES shop (id),
  path_params text[] NOT NULL,
  query_params text[] NOT NULL,
  description text NOT NULL, -- template
  image text, -- data-uri or nothing
  currency text NOT NULL DEFAULT 'sat', -- sat, usd, eur, brl etc.
  min_price text NOT NULL, -- formula
  max_price text NOT NULL, -- formula

  PRIMARY KEY (shop, id),
  CONSTRAINT params_overlap CHECK (not (path_params && query_params)),
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
  shop text NOT NULL,
  template text NOT NULL,
  params jsonb NOT NULL,
  creation timestamp NOT NULL DEFAULT now(),
  payment timestamp, -- null when not paid
  amount_msat numeric(13) NOT NULL,
  bolt11 text NOT NULL,

  FOREIGN KEY (shop, template) REFERENCES template (shop, id)
);
