package state

const CreateTablesSQL = `
--
-- Definitions
--

CREATE TABLE IF NOT EXISTS task_def (
  definition_id character varying PRIMARY KEY,
  alias character varying,
  image character varying NOT NULL,
  group_name character varying NOT NULL,
  memory integer,
  command text,
  -- Refactor these
  "user" character varying,
  arn character varying,
  container_name character varying NOT NULL,
  task_type character varying,
  -- Refactor these
  CONSTRAINT task_def_alias UNIQUE(alias)
);

CREATE TABLE IF NOT EXISTS task_def_environments (
  task_def_id character varying NOT NULL REFERENCES task_def(definition_id),
  name character varying NOT NULL,
  value character varying,
  CONSTRAINT task_def_environments_pkey PRIMARY KEY(task_def_id, name)
);

CREATE TABLE IF NOT EXISTS task_def_ports (
  task_def_id character varying NOT NULL REFERENCES task_def(definition_id),
  port integer NOT NULL,
  CONSTRAINT task_def_ports_pkey PRIMARY KEY(task_def_id, port)
);

CREATE INDEX IF NOT EXISTS ix_task_def_alias ON task_def(alias);
CREATE INDEX IF NOT EXISTS ix_task_def_group_name ON task_def(group_name);
CREATE INDEX IF NOT EXISTS ix_task_def_image ON task_def(image);

--
-- Runs
--

CREATE TABLE IF NOT EXISTS task (
  run_id character varying NOT NULL PRIMARY KEY,
  definition_id character varying REFERENCES task_def(definition_id),
  cluster_name character varying,
  exit_code integer,
  status character varying,
  started_at timestamp with time zone,
  finished_at timestamp with time zone,
  instance_id character varying,
  instance_dns_name character varying,
  group_name character varying,
  -- Refactor these --
  task_arn character varying,
  docker_id character varying,
  "user" character varying,
  task_type character varying
  -- Refactor these --
);

CREATE TABLE IF NOT EXISTS task_environments (
  task_id character varying NOT NULL REFERENCES task(run_id),
  name character varying NOT NULL,
  value character varying,
  CONSTRAINT task_environments_pkey PRIMARY KEY(task_id, name)
);

CREATE INDEX IF NOT EXISTS ix_task_cluster_name ON task(cluster_name);
CREATE INDEX IF NOT EXISTS ix_task_status ON task(status);
CREATE INDEX IF NOT EXISTS ix_task_group_name ON task(group_name);

--
-- Status
--

CREATE TABLE IF NOT EXISTS task_status (
  status_id integer NOT NULL PRIMARY KEY,
  task_arn character varying,
  status_version integer NOT NULL,
  status character varying,
  "timestamp" timestamp with time zone DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_task_status_task_arn ON task_status(task_arn);

CREATE SEQUENCE IF NOT EXISTS task_status_status_id_seq
  START WITH 1
  INCREMENT BY 1
  NO MINVALUE
  NO MAXVALUE
  CACHE 1;

ALTER TABLE ONLY task_status ALTER COLUMN status_id SET DEFAULT nextval('task_status_status_id_seq'::regclass);
`
const DefinitionSelect = `
select
  coalesce(td.arn,'')       as arn,
  td.definition_id          as definitionid,
  td.image                  as image,
  td.group_name             as groupname,
  td.container_name         as containername,
  coalesce(td.user,'')      as "user",
  td.alias                  as alias,
  td.memory                 as memory,
  coalesce(td.command,'')   as command,
  coalesce(td.task_type,'') as tasktype,
  env::TEXT                 as env,
  ports                     as ports
  from (select * from task_def) td left outer join
    (select task_def_id,
      array_to_json(
        array_agg(json_build_object('name',name,'value',coalesce(value,'')))) as env
          from task_def_environments group by task_def_id
    ) tde
  on td.definition_id = tde.task_def_id left outer join
    (select task_def_id,
      array_to_json(array_agg(port))::TEXT as ports
        from task_def_ports group by task_def_id
    ) tdp
  on td.definition_id = tdp.task_def_id
`

const ListDefinitionsSQL = DefinitionSelect + "\n%s %s limit $1 offset $2"
const GetDefinitionSQL = DefinitionSelect + "\nwhere definition_id = $1"

const RunSelect = `
select
  coalesce(t.task_arn,'')                    as taskarn,
  t.run_id                                   as runid,
  coalesce(t.definition_id,'')               as definitionid,
  coalesce(t.cluster_name,'')                as clustername,
  t.exit_code                                as exitcode,
  coalesce(t.status,'')                      as status,
  started_at                                 as startedat,
  finished_at                                as finishedat,
  coalesce(t.instance_id,'')                 as instanceid,
  coalesce(t.instance_dns_name,'')           as instancednsname,
  coalesce(t.group_name,'')                  as groupname,
  coalesce(t.user,'')                        as "user",
  coalesce(t.task_type,'')                   as tasktype,
  env::TEXT                                  as env
  from (select * from task) t left outer join
    (select task_id,
      array_to_json(
        array_agg(json_build_object('name',name,'value',coalesce(value,'')))) as env
          from task_environments group by task_id
    ) te
  on t.run_id = te.task_id
`
const ListRunsSQL = RunSelect + "\n%s %s limit $1 offset $2"
const GetRunSQL = RunSelect + "\nwhere run_id = $1"
