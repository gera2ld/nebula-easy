export interface ITransactionParams {
  cwd: string;
}

export interface INebulaHost {
  type: "lighthouse" | "host";
  name: string;
  ip: string;
  relay: boolean;
  publicIpPort: string;
}

export interface INebulaNetwork {
  name: string;
  ipRange: string;
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
