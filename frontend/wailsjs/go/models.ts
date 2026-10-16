export namespace model {
	
	export class ServiceConfig {
	    id: string;
	    name: string;
	    exePath: string;
	    workDir: string;
	    args: string[];
	    env: Record<string, string>;
	    logFilePath: string;
	    logEncoding: string;
	    autoRestart: boolean;
	    startOrder: number;
	    startDelay: number;
	
	    static createFrom(source: any = {}) {
	        return new ServiceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.exePath = source["exePath"];
	        this.workDir = source["workDir"];
	        this.args = source["args"];
	        this.env = source["env"];
	        this.logFilePath = source["logFilePath"];
	        this.logEncoding = source["logEncoding"];
	        this.autoRestart = source["autoRestart"];
	        this.startOrder = source["startOrder"];
	        this.startDelay = source["startDelay"];
	    }
	}
	export class ServiceInfo {
	    config: ServiceConfig;
	    status: string;
	    pid: number;
	    startTime: number;
	    exitCode: number;
	    lastLog: string;
	
	    static createFrom(source: any = {}) {
	        return new ServiceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config = this.convertValues(source["config"], ServiceConfig);
	        this.status = source["status"];
	        this.pid = source["pid"];
	        this.startTime = source["startTime"];
	        this.exitCode = source["exitCode"];
	        this.lastLog = source["lastLog"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

