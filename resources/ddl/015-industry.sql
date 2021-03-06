DROP TABLE IF EXISTS app.industry_jobs;
CREATE TABLE app.industry_jobs
(
  job_id BIGINT PRIMARY KEY NOT NULL,
  corporation_id BIGINT NOT NULL,
  installer_id BIGINT NOT NULL,
  facility_id BIGINT NOT NULL,
  location_id BIGINT NOT NULL,
  activity_id BIGINT NOT NULL,
  blueprint_id BIGINT NOT NULL,
  blueprint_type_id INT NOT NULL,
  blueprint_location_id BIGINT NOT NULL,
  output_location_id BIGINT NOT NULL,
  product_type_id INT NOT NULL,
  runs INT NOT NULL,
  cost NUMERIC NOT NULL,
  licensed_runs INT NOT NULL,
  probability NUMERIC NOT NULL,
  status VARCHAR(100) NOT NULL,
  start_date TIMESTAMP NOT NULL,
  end_date TIMESTAMP NOT NULL,
  pause_date TIMESTAMP NOT NULL,
  completed_date TIMESTAMP NOT NULL,
  completed_character_id BIGINT NOT NULL,
  successful_runs INT NOT NULL,
  fetched_at TIMESTAMP NOT NULL DEFAULT NOW()
);
