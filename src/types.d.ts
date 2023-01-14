interface ITransactionParams {
  cwd: string;
}

interface INebulaHost {
  type: "lighthouse" | "host";
  name: string;
  ip: string;
  relay: boolean;
}

interface INebulaNetwork {
  name: string;
  ip: string;
  staticHostMap: Record<string, string[]>;
  hosts: INebulaHost[];
}

interface INebulaData {
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
