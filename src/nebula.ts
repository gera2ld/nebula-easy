import { runCommand } from "https://raw.githubusercontent.com/gera2ld/deno-lib/main/lib/cli.ts";
import { limitConcurrency } from "https://raw.githubusercontent.com/gera2ld/deno-lib/main/lib/util.ts";
import type { INebulaData, ITransactionParams } from "./types.ts";

const nebulaCert = Deno.env.get("NEBULA_CERT") || "nebula-cert";
const dataPath = Deno.env.get('DATA_PATH') || 'data/db.json';

let nebulaData: INebulaData = {
  secrets: {},
  networks: [],
};

let id = 0;
const limitedTransaction = limitConcurrency(transaction, 1);

async function transaction<T>(
  callback: (params: ITransactionParams) => Promise<T>,
) {
  id += 1;
  const cwd = `/tmp/nebula_${id}`;
  await Deno.mkdir(cwd, { recursive: true });
  try {
    return await callback({ cwd });
  } finally {
    await Deno.remove(cwd, { recursive: true });
  }
}

export function createCA(name: string) {
  return limitedTransaction(async ({ cwd }) => {
    const cmd = [nebulaCert, "ca", "-name", name];
    console.info("$", cmd.join(" "));
    await runCommand({
      cmd,
      cwd,
    });
    const [key, crt] = await Promise.all([
      Deno.readTextFile(`${cwd}/ca.key`),
      Deno.readTextFile(`${cwd}/ca.crt`),
    ]);
    nebulaData.ca = { name, crt };
    nebulaData.secrets = { ca: { key } };
    dumpData();
    return { crt };
  });
}

export function signCert(params: {
  name: string;
  ip: string;
  pub?: string;
}) {
  return limitedTransaction(async ({ cwd }) => {
    if (!nebulaData.ca || !nebulaData.secrets.ca) {
      throw new Error("CA not found");
    }
    await Deno.writeTextFile(`${cwd}/ca.key`, nebulaData.secrets.ca.key);
    await Deno.writeTextFile(`${cwd}/ca.crt`, nebulaData.ca.crt);
    if (params.pub) await Deno.writeTextFile(`${cwd}/host.pub`, params.pub);
    const cmd = [
      nebulaCert,
      "sign",
      "-name",
      params.name,
      "-ip",
      params.ip,
      ...params.pub
        ? [
          "-in-pub",
          "host.pub",
        ]
        : [],
      "-out-key",
      "host.key",
      "-out-crt",
      "host.crt",
    ];
    console.info("$", cmd.join(" "));
    await runCommand({
      cmd,
      cwd,
    });
    const [crt, key] = await Promise.all([
      Deno.readTextFile(`${cwd}/host.crt`),
      params.pub ? null : Deno.readTextFile(`${cwd}/host.key`),
    ]);
    return { key, crt };
  });
}

function getRelayConfig(relays: string[] | boolean) {
  if (relays === true) {
    return {
      am_relay: true,
      use_relays: false,
    };
  }
  return {
    am_relay: false,
    use_relays: true,
    relays: relays || [],
  };
}

const baseConfig = {
  pki: {
    ca: "/etc/nebula/ca.crt",
    cert: "/etc/nebula/host.crt",
    key: "/etc/nebula/host.key",
  },
  lighthouse: {
    am_lighthouse: true,
  },
  listen: {
    host: "0.0.0.0",
    port: 4242,
  },
};

export function getLighthouseConfig(relays = true) {
  const data = {
    ...baseConfig,
    relay: getRelayConfig(relays),
  };
  return data;
}

export function getHostConfig(params: {
  staticHostMap: Record<string, string[]>;
  lighthouseHosts: string[];
  relays: string[] | boolean;
}) {
  const data = {
    ...baseConfig,
    lighthouse: {
      am_lighthouse: false,
      hosts: params.lighthouseHosts,
    },
    static_host_map: params.staticHostMap,
    relay: getRelayConfig(params.relays),
  };
  return data;
}

export async function loadData(path = dataPath) {
  try {
    nebulaData = JSON.parse(await Deno.readTextFile(path));
  } catch {
    // ignore
  }
  return {
    ...nebulaData,
    secrets: undefined,
  };
}

export async function dumpData(data?: INebulaData, path = dataPath) {
  console.info("> Dump data");
  nebulaData = {
    ...nebulaData,
  };
  if (data?.networks) nebulaData.networks = data.networks;
  await Deno.writeTextFile(path, JSON.stringify(nebulaData));
}
