import fs from "fs";

readFileAndParseToJsonArray();

function readFileAndParseToJsonArray(): void {
  const inputFile = process.argv[2];
  const outputFile = process.argv[3];

  if (typeof inputFile !== "string") {
    console.error("Please provide input file path as first argument.");
    return;
  }
  if (typeof outputFile !== "string") {
    console.error("Please provide output file path as second argument.");
    return;
  }

  const jsonArray: any[] = [];

  try {
    const fileContents: string = fs.readFileSync(inputFile, "utf-8");
    const lines: string[] = fileContents.split("\n");

    for (const line of lines) {
      try {
        const jsonData = JSON.parse(line);
        jsonData.msg = JSON.parse(jsonData.msg);
        jsonArray.push(jsonData);
      } catch (error) {
        //   do nothing
      }
    }


    const arr = jsonArray.filter((item) => {
      return item.requestId;
    });

    const arrOrdered = arr.map((item, index) => {
      return { ...item, orderNumber: index };
    });

    const logByRequestId: any = {};
    arrOrdered.forEach((item) => {
      const requestId: string = item.requestId;
      if (item.requestDirection === "incoming") {
        logByRequestId[requestId] = item;
      } else if (item.requestDirection === "outgoing") {
        const current = logByRequestId[requestId];
        if (current?.msg) {
          logByRequestId[requestId].msg = { ...current.msg, ...item.msg };
        }
      }
    });

    const arrCombined = Object.values(logByRequestId).filter((item: any) => {
      return item?.msg?.AdmissionReview;
    });

    arrCombined.sort((a: any, b: any) => {
      return a.orderNumber - b.orderNumber;
    });

    arrCombined.forEach((item: any) => {
      delete item.orderNumber;
    });

    // fs.writeFileSync('raw-'+outputFile, JSON.stringify(arrCombined, null, 2), "utf-8");

    const analyzedLogs = analyzeLogs(arrCombined as any);

    const jsonArrayString: string = JSON.stringify(analyzedLogs, null, 2);

    fs.writeFileSync(outputFile, jsonArrayString, "utf-8");
    console.log("File parsing and writing completed successfully.");
  } catch (error) {
    console.error("Error occurred while reading or writing the file:", error);
  }
}

type OwnerReference = {
  apiVersion: string;
  kind: string;
  name: string;
  uid: string;
};

type Metadata = {
  name: string;
  namespace: string;
  uid: string;
  resourceVersion: string;
  creationTimestamp: string;
  ownerReferences: OwnerReference[];
  managedFields: {
    manager: string;
    operation: string;
    apiVersion: string;
    time: string;
    fieldsType: string;
    fieldsV1: {
      [key: string]: object;
    };
  }[];
};

type ObjectSpec = {
  holderIdentity: string;
  leaseDurationSeconds: number;
  renewTime: string;
};

type ObjectData = {
  kind: string;
  apiVersion: string;
  metadata: Metadata;
  spec: ObjectSpec;
};

type AdmissionReviewRequest = {
  uid: string;
  kind: {
    group: string;
    version: string;
    kind: string;
  };
  resource: {
    group: string;
    version: string;
    resource: string;
  };
  requestKind: {
    group: string;
    version: string;
    kind: string;
  };
  requestResource: {
    group: string;
    version: string;
    resource: string;
  };
  name: string;
  namespace: string;
  operation: string;
  userInfo: {
    username: string;
    groups: string[];
  };
  object: ObjectData;
  oldObject: ObjectData;
  dryRun: boolean;
  options: {
    kind: string;
    apiVersion: string;
  };
};

type AdmissionReviewResponse = {
  uid: string;
  allowed: boolean;
  status: {
    metadata: object;
    message: string;
    code: number;
  };
};

type JSONData = {
  level: string;
  ts: number;
  caller: string;
  msg: {
    kind: string;
    apiVersion: string;
    request: AdmissionReviewRequest;
    AdmissionReview: {
      kind: string;
      apiVersion: string;
      response: AdmissionReviewResponse;
    };
    IsSkipped: boolean;
  };
  requestId: string;
  requestDirection: string;
};

function analyzeLogs(logs: JSONData[]): any[] {
  return logs.map((log) => ({
    isSkipped: log.msg.IsSkipped,
    isAllowed: log.msg.AdmissionReview.response.allowed,
    kind: log.msg.request.object.kind,
    name: log.msg.request.object.metadata.name,
    ownerReferences: log.msg.request.object.metadata.ownerReferences,
    managedFields: log.msg.request.object.metadata.managedFields,
    namespace: log.msg.request.namespace,
    userInfo: log.msg.request.userInfo,
    operation: log.msg.request.operation,
    dryRun: log.msg.request.dryRun
  })).filter((log) => {
    return log.kind !== "Lease"
      && log.kind !== "Event"
      && log.kind !== "SubjectAccessReview"
      && !log?.namespace?.startsWith("openshift")
      && log.userInfo.username.startsWith("system:")
      && !log.userInfo.username.startsWith("system:serviceaccount")
      && !log.userInfo.username.startsWith("system:node");
  });
}
