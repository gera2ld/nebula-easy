export interface ITransactionParams {
  cwd: string;
}

export interface INebulaHost {
  type: "lighthouse" | "host";
  name: string;
  ip: string;
  relay: boolean;
  publicIp: string;
}

export interface INebulaNetwork {
  name: string;
  ip: string;
  staticHostMap: Record<string, string[]>;
  hosts: INebulaHost[];
}

export interface INebulaData {
  ca?: {
    name: string;
    crt: string;
  };
  secrets: {
    ca?: {
      key: string;
    };
  };
  networks: INebulaNetwork[];
}
