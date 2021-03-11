/**
 * ------------------------------------------------------------------
 * Native definitions.
 * ------------------------------------------------------------------
 */

// for the ts compiler
interface Boolean { }
interface Function { }
interface IArguments { }
interface Number { }
interface Object { }
interface RegExp { }
interface byte { }

declare const int: any
declare const float: any
declare const Array: any

interface Array<T> {
    [n: number]: T
    slice(start?: number, count?: number): Array<T>
    range(start?: number, end?: number): Array<T>
    append(v: T[]): T[]
    push(...v: T[]): void
    pushRange(v: T[]): void
    length: number
    insertAt(i: number, v: T): void
    removeAt(i: number): void
    removeAt(from: number, to: number): void
    indexOf(v: T): number
    join(sep: string): T
    sort(comprarer: (a: T, b: T) => boolean): void
}

// translate a value.
declare function T(key: string, ...params: any[]): string

declare namespace errors {
    export function wrap(msg: string, inner: Error): Error
    export function public(msg: string, inner?: Error | string): Error

    export interface Error {
        type: string
        public: boolean
        message: string
        pc: number
        stackTrace: string
        toString(): string
    }
}





interface Array<T> {
    [n: number]: T
    slice(start?: number, count?: number): Array<T>
    range(start?: number, end?: number): Array<T>
    append(v: T[]): T[]
    push(...v: T[]): void
    pushRange(v: T[]): void
    copyAt(i: number, v: T[]): void
    length: number
    insertAt(i: number, v: T): void
    removeAt(i: number): void
    removeAt(from: number, to: number): void
    indexOf(v: T): number
    join(sep: string): T
    sort(comprarer: (a: T, b: T) => boolean): void
    equals(other: Array<T>): boolean;
    any(func: (t: T) => any): boolean;
    all(func: (t: T) => any): boolean;
    contains<T>(t: T): boolean;
    remove<T>(t: T): void;
    first(): T;
    last(): T;
    first(func?: (t: T) => any): T;
    last(func?: (t: T) => any): T;
    firstIndex(func: (t: T) => any): number;
    select<K>(func: (t: T) => K): Array<K>;
    selectMany<K>(func: (t: T) => K): K;
    distinct<K>(func?: (t: any) => K): Array<K>;
    where(func: (t: T) => any): Array<T>;
    groupBy(func: (t: T) => any): KeyIndexer<T[]>;
    sum<K extends number>(): number;
    sum<K extends number>(func: (t: T) => K): number;
    min(func: (t: T) => number): number;
    max(func: (t: T) => number): number;
    count(func: (t: T) => any): number;
}

declare namespace array {
    /**
     * Create a new array with size.
     */
    export function make<T>(size: number, capacity?: number): Array<T>

    /**
     * Create a new array of bytes with size.
     */
    export function bytes(size: number, capacity?: number): byte[]
}




declare namespace assert {
    export function contains(value: string, search: string): void
    export function equal(a: any, b: any): void
    export function isNull(a: any): void
    export function isNotNull(a: any): void
    export function exception(msg: string, func: Function): void

    export function int(a: any, msg: string): number
    export function float(a: any, msg: string): number
    export function string(a: any, msg: string): string
    export function bool(a: any, msg: string): boolean
    export function object(a: any, msg: string): any
}





declare function go(f: Function): void






declare namespace base64 {
    export function encode(s: any): string
    export function encodeWithPadding(s: any): string
    export function decode(s: any): string
    export function decodeWithPadding(s: any): string
}





declare namespace binary {
    export function putInt16LittleEndian(v: byte[], n: number): void
    export function putInt32LittleEndian(v: byte[], n: number): void
    export function putInt16BigEndian(v: byte[], n: number): void
    export function putInt32BigEndian(v: byte[], n: number): void

    export function int16LittleEndian(v: byte[]): number
    export function int32LittleEndian(v: byte[]): number
    export function int16BigEndian(v: byte[]): number
    export function int32BigEndian(v: byte[]): number
    export function int64BigEndian(v: byte[]): number
}





declare namespace bufio {
    export function newWriter(w: io.Writer): Writer
    export function newScanner(r: io.Reader): Scanner
    export function newReader(r: io.Reader): Reader

    export interface Writer {
        write(data: byte[]): number
        writeString(s: string): number
        writeByte(b: byte): void
        writeRune(s: string): number
        flush(): void
    }

    export interface Scanner {
        scan(): boolean
        text(): string
    }

    export interface Reader {
        readString(delim: byte): string
        readBytes(delim: byte): byte[]
        readByte(): byte
        readRune(): number
    }
}





declare namespace bytecode {
    /**
     * 
     * @param path 
     * @param fileSystem 
     * @param scriptMode if statements outside of functions are allowed.
     */
    export function compile(path: string, fileSystem?: io.FileSystem): runtime.Program

    export function hash(path: string, fileSystem?: io.FileSystem): string

    export function compileStr(code: string): runtime.Program

    /**
     * Load a binary program from the file system
     * @param path the path to the main binary.
     * @param fs the file trusted. If empty it will use the current fs.
     */
    export function load(path: string, fs?: io.FileSystem): runtime.Program

    export function loadProgram(b: byte[]): runtime.Program
    export function readProgram(r: io.Reader): runtime.Program
    export function writeProgram(w: io.Writer, p: runtime.Program): void
}





declare namespace caching {

    export let global: Cache

    export function newCache(d?: time.Duration | number): Cache

    export interface Cache {
        get(key: string): any | null
        save(key: string, v: any): void
        delete(key: string): void
        keys(): string[]
        items(): Map<any>
        clear(): void
    }
}





declare namespace console {
    export function log(...v: any[]): void
}




declare namespace convert {
    export function toInt(v: string | number): number
    export function toFloat(v: string | number): number
    export function parseCurrency(v: string | number): number
    export function toString(v: any): string
    export function toRune(v: any): string
    export function toBool(v: string | number | boolean): boolean
    export function toBytes(v: string | byte[]): byte[]
}





declare namespace crypto {
    export function signSHA1_RSA_PCKS1(privateKey: string, value: string): byte[]
    // export function verifySHA1_RSA(publicKey: string, message: string, signature: byte[]): void

    export function signTempSHA1(value: string): string
    export function checkTempSignSHA1(value: string, hash: string): boolean

    export function signSHA1(value: string): string
    export function checkSignSHA1(value: string, hash: string): boolean

    export function setGlobalPassword(pwd: string): void
    export function encrypt(value: byte[], pwd?: byte[]): byte[]
    export function decrypt(value: byte[], pwd?: byte[]): byte[]
    export function encryptTripleDES(value: byte[] | string, pwd?: byte[] | string): byte[]
    export function decryptTripleDES(value: byte[] | string, pwd?: byte[] | string): byte[]
    export function encryptString(value: string, pwd?: string): string
    export function decryptString(value: string, pwd?: string): string
    export function hashSHA(value: string): string
    export function hashSHA256(value: string): string
    export function hashSHA512(value: string): string
    export function hmacSHA256(value: byte[] | string, pwd?: byte[] | string): byte[]
    export function hmacSHA512(value: byte[] | string, pwd?: byte[] | string): byte[]
    export function hashPassword(pwd: string): string
    export function compareHashAndPassword(hash: string, pwd: string): boolean
    export function random(len: number): byte[]
    export function randomAlphanumeric(len: number): string
}






declare namespace csv {
    export function newReader(r: io.Reader): Reader
    export interface Reader {
        comma: string
        lazyQuotes: boolean
        read(): string[]
    }

    export function newWriter(r: io.Writer): Writer
    export interface Writer {
        comma: string
        write(v: (string | number)[]): void
        flush(): void
    }
}




declare namespace encoding {
    export interface Decoder {
        reader(r: io.Reader): io.Reader
    }
    export interface Encoder {
        writer(r: io.Writer): io.Writer
        string(s: string): string
    }

    export function newDecoderISO8859_1(): Decoder
    export function newEncoderISO8859_1(): Encoder
    export function newDecoderWindows1252(): Decoder
    export function newEncoderWindows1252(): Encoder
    export function newDecoderUTF16_LE(): Decoder
    export function newEncoderUTF16_LE(): Encoder
}





declare namespace filepath {
    /**
     * Returns the extension of a path
     */
    export function ext(path: string): string

    /**
     *  Base returns the last element of path.
     *  Trailing path separators are removed before extracting the last element.
     *  If the path is empty, Base returns ".".
     *  If the path consists entirely of separators, Base returns a single separator.
     */
    export function base(path: string): string

    /**
     * Returns name of the file without the directory and extension.
     */
    export function nameWithoutExt(path: string): string

    /**
     * Returns directory part of the path.
     */
    export function dir(path: string): string

    export function join(...parts: string[]): string

    /**
     * joins the elemeents but respecting absolute paths.
     */
    export function joinAbs(...parts: string[]): string
}





declare namespace fileutil {
    export function copy(src: string, dst: string): byte[]
}




declare namespace fmt {
    export function print(...n: any[]): void
    export function println(...n: any[]): void
    export function printf(format: string, ...params: any[]): void
    export function sprintf(format: string, ...params: any[]): string
    export function fprintf(w: io.Writer, format: string, ...params: any[]): void
}




declare namespace fsnotify {
    export function newWatcher(onEvent: EventHandler): Watcher

    export type EventHandler = (e: Event) => void

    export interface Watcher {
        add(path: string, recursive?: boolean): void
    }

    export interface Event {
        name: string
        operation: number
    }

    // const (
    // 	Create Op = 1 << iota
    // 	Write
    // 	Remove
    // 	Rename
    // 	Chmod
    // )
}





declare namespace gzip {
    export function newWriter(w: io.Writer): io.WriterCloser
    export function newReader(r: io.Reader): io.ReaderCloser
}






declare namespace hash {
    export function newMD5(): Hash
    export function newSHA256(): Hash

    export interface Hash {
        write(b: byte[]): number
        sum(): byte[]
    }
}






declare namespace hex {
    export function encodeToString(b: byte[]): number
}






declare namespace html {
    export function encode(s: any): string
    export function decode(s: any): string
}







declare namespace http {
    export type SameSite = number
    export const SameSiteDefaultMode: SameSite
    export const SameSiteLaxMode: SameSite
    export const SameSiteStrictMode: SameSite
    export const SameSiteNoneMode: SameSite

    export function get(url: string, timeout?: time.Duration | number, config?: tls.Config): string
    export function post(url: string, data?: any): string

    export function getJSON(url: string): any

    export function cacheBreaker(): string
    export function resetCacheBreaker(): string

    export function encodeURIComponent(url: string): string
    export function decodeURIComponent(url: string): string

    export function parseURL(url?: string): URL

    export function newRouter(): Router

    export interface Router {
        reset(): void
        add(r: Route): void
        match(url: string): RouteMatch | null
    }

    export interface RouteMatch {
        route: Route
        values: string[]
        int(name: string): number
        value(name: string): string
    }

    export interface Route {
        url: string
        handler: Function
        filter?: Function
        permissions?: string[]
    }

    export type Handler = (w: ResponseWriter, r: Request) => void

    export interface Server {
        address: string
        addressTLS: string
        tlsConfig: tls.Config
        handler: Handler
        readHeaderTimeout: time.Duration | number
        writeTimeout: time.Duration | number
        readTimeout: time.Duration | number
        idleTimeout: time.Duration | number
        start(): void
        close(): void
        shutdown(duration?: time.Duration | number): void
    }

    export function newServer(): Server

    export type METHOD = "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "OPTIONS"

    export function newRequest(method: METHOD, url: string, data?: any): Request

    export function newContext(method: METHOD, url: string, data?: any): Context

    export interface Context {
        request: Request
        response: ResponseWriter
    }

    export interface Request {
        /**
         * If the request is using a TLS connection
         */
        tls: boolean

        /**
         * The http method.
         */
        method: METHOD

        host: string

        /**
         * scheme + host + port
         * https://stackoverflow.com/a/37366696/4264
         */
        origin: string

        url: URL

        referer: string

        userAgent: string

        body: io.ReaderCloser

        remoteAddr: string
        remoteIP: string

        /**
         * The extension of the URL
         */
        extension: string

        // string returns the first value for the named component of the query.
        // POST and PUT body parameters take precedence over URL query string values.
        string(key: string): string

        // int works as value but deserializes the value into an object.
        json(key: string): any

        // int works as value but converts the value to an int.
        int(key: string): number

        // float works as value but converts the value to a float.
        float(key: string): number

        // currency works as value but converts the value to a float.
        currency(key: string): number

        // bool works as value but converts the value to an bool.
        bool(key: string): boolean

        // date works as value but converts the value formated as "yyyy-mm-dd" to an time.Time.
        date(key: string): time.Time

        headers(): string[]
        header(key: string): string
        setHeader(key: string, value: string): void

        file(name: string): File

        // The parsed form data, including both the URL
        // field's query parameters and the POST or PUT form data.
        // This field is only available after ParseForm is called.
        // The HTTP client ignores Form and uses Body instead.
        values: StringMap

        formValues: StringMap

        cookie(key: string): Cookie | null

        addCookie(c: Cookie): void
        addCookieValue(name: string, value: string): void

        setBasicAuth(user: string, password: string): void
        basicAuth(): { user: string, password: string }

        execute(timeout?: number | time.Duration, tlsconf?: tls.Config): Response

        mustExecute(timeout?: number | time.Duration, tlsconf?: tls.Config): string
    }


    export interface File {
        name: string
        contentType: string
        size: number
        read(b: byte[]): number
        ReadAt(p: byte[], off: number): number
        close(): void
    }

    export function newCookie(): Cookie

    export interface Cookie {
        domain: string
        path: string
        expires: time.Time
        name: string
        value: string
        secure: boolean
        httpOnly: boolean
        sameSite: SameSite
    }

    export interface URL {
        scheme: string
        host: string
        port: string

        /**
         * The host without the port number if present
         */
        hostName: string

        /**
         * returns the subdomain part *only if* the host has a format xxx.xxxx.xx.
         */
        subdomain: string

        path: string
        query: string
        pathAndQuery: string
    }

    // interface FormValues {
    //     [key: string]: any
    // }  


    export interface Response {
        status: number
        handled: boolean
        proto: string
        body(): string
        bytes(): byte[]
        cookies(): Cookie[]
        headers(): string[]
        header(name: string): string[]
    }


    export interface ResponseWriter {
        readonly status: number

        handled: boolean

        /**
         * Only will have a value if it is a Recorder
         */
        readonly body: io.Buffer

        readonly headers: Map<string[]>

        cookie(name: string): Cookie

        cookies(): Cookie[]

        addCookie(c: Cookie): void

        /**
         * Writes v to the server response.
         */
        write(v: any): number

        /**
         * Writes v to the server response setting json content type if
         * the header is not already set.
         */
        writeJSON(v: any, skipCacheHeader?: boolean): void

        /**
         * Writes v to the server response setting json content type if
         * the header is not already set.
         */
        writeJSONStatus(status: number, v: any, skipCacheHeader?: boolean): void

        /**
         * Serves a static file
         */
        writeFile(name: string, data: byte[] | string | io.File | io.FileSystem): void

        /**
         * Sets the http status header.
         */
        setStatus(status: number): void

        /**
         * Sets the content type header.
         */
        setContentType(type: string): void

        /**
         * Sets the content type header.
         */
        setHeader(name: string, value: string): void

        /**
         * Send a error to the client
         */
        writeError(status: number, msg?: string): void

        /**
         * Send a error with json content-type to the client
         */
        writeJSONError(status: number, msg?: string): void

        redirect(url: string): void
    }


}





declare namespace httputil {
    export function newSingleHostReverseProxy(target: http.URL): ReverseProxy

    export interface ReverseProxy {
        serveHTTP(w: http.ResponseWriter, r: http.Request): void
    }
}





declare namespace i18n {
    export interface Culture {
        name: string
        currencySymbol: string
        currency: string
        currencyPattern: string
        decimalSeparator: string
        thousandSeparator: string
        shortDatePattern: string
        longDatePattern: string
        shortTimePattern: string
        longTimePattern: string
        dateTimePattern: string
        firstDayOfWeek: number
    }

    export const DefaultCulture: Culture
    export function setDefaultCulture(c: string): void

    export function getCultureNames(): StringMap
    export function getCulture(name: string): Culture
    export function format(pattern: string, value: any): string
    export function translate(culture: string, key: string): string
    export function addTranslation(culture: string, key: string, translation: string): void
    export function addResources(culture: string, name: string, resources: any[]): void



    /**
     *  Example:
            "es-ES",
            ',',
            '.',
            "EUR",
            "€",
            "0:00€",
            "0:0000€",
            "0:00",
            "dd-MM-yyyy HH:mm",
            "ddd, dd-MM-yyyy HH:mm",
            "dd-MM-yyyy",
            "ddd, dd MMM yyyy",
            "HH:mm",
            "HH:mm:ss",
            time.Monday,
     */
    export function addCulture(
        name: string,
        decimalSeparator: string,
        thousandSeparator: string,
        currency: string,
        currencySymbol: string,
        currencyPattern: string,
        floatPattern: string,
        dateTimePattern: string,
        shortDatePattern: string,
        longDatePattern: string,
        shortTimePattern: string,
        longTimePattern: string,
        firstDayOfWeek: number): Culture
}




declare namespace io {
    export interface Reader {
        read(b: byte[]): number
    }

    export interface ReaderAt {
        ReadAt(p: byte[], off: number): number
    }

    export interface ReaderCloser extends Reader {
        close(): void
    }

    export interface Writer {
        write(v: string | byte[]): number | void
    }

    export interface WriterCloser extends Writer {
        close(): void
    }

    export function copy(dst: Writer, src: Reader): number

    export function newMemFS(): FileSystem

    export function newRootedFS(root: string, baseFS: FileSystem): FileSystem

    export function newRestrictedFS(baseFS: FileSystem): RestrictedFS

    /** 
     * Sets the default data file system that will be returned by io.dataFS()
     */
    export function setDataFS(fs: FileSystem): void

    export function newBuffer(): Buffer

    export interface Buffer {
        length: number
        cap: number
        read(b: byte[]): number
        write(v: any): void
        toString(): string
        toBytes(): byte[]
    }

    export interface FileSystem {
        /**
         * The current working directory
         */
        workingDir: string

        abs(path: string): string
        open(path: string): File
        openIfExists(path: string): File
        openForWrite(path: string): File
        openForAppend(path: string): File
        chdir(dir: string): void
        exists(path: string): boolean
        rename(source: string, dest: string): void
        removeAll(path: string): void
        readAll(path: string): byte[]
        readAllIfExists(path: string): byte[]
        readString(path: string): string
        readStringIfExists(path: string): string
        write(path: string, data: string | io.Reader | byte[]): void
        append(path: string, data: string | byte[]): void
        mkdir(path: string): void
        stat(path: string): FileInfo
        readDir(path: string): FileInfo[]
        readNames(path: string, recursive?: boolean): string[]
    }

    export interface RestrictedFS extends FileSystem {
        addToWhitelist(path: string): void
        addToBlacklist(path: string): void
    }

    export interface File {
        read(b: byte[]): number
        write(v: string | byte[] | io.Reader): number
        writeAt(v: string | byte[] | io.Reader, offset: number): number
        close(): void
    }

    export interface FileInfo {
        name: string
        modTime: time.Time
        isDir: boolean
        size: number
    }
}




declare namespace ioutil {
    export function readAll(r: io.Reader): byte[]
}





declare namespace json {
    export function escapeString(str: string): string
    export function marshal(v: any, indent?: boolean): string
    export function unmarshal(str: string | byte[]): any

}




declare namespace log {
    export const defaultLogger: Logger
    export function setDefaultLogger(logger: Logger): void
    export function fatal(format: any, ...v: any[]): void
    export function system(format: any, ...v: any[]): void
    export function write(table: string, format: any, ...v: any[]): void

    export function newLogger(path: string, fs?: io.FileSystem): Logger

    export interface Logger {
        path: string
        debug: boolean
        save(table: string, data: string, ...v: any): void
        insert(date: time.Time, table: string, data: string, ...v: any[]): void
        query(table: string, start: time.Time, end: time.Time, offset?: number, limit?: number): Scanner
    }

    export interface Scanner {
        scan(): boolean
        data(): DataPoint
        setFilter(v: string): void
    }

    export interface DataPoint {
        text: string
        time: time.Time
        string(): string
    }
}



declare interface StringMap {
    [key: string]: string
}

declare interface KeyIndexer<T> {
    [key: string]: T
}

declare type Map<T> = KeyIndexer<T>

declare namespace Object {
    export function len(v: any): number
    export function keys(v: any): string[]
    export function values<T>(v: Map<T>): T[]
    export function values<T>(v: any): T[]
    export function deleteKey(v: any, key: string | number): void
    export function deleteKeys(v: any): void
    export function hasKey(v: any, key: any): boolean
    export function clone<T>(v: T): T
}




declare namespace markdown {

    export function toHTML(n: string | byte[]): string
}





declare namespace math {
    /**
     * returns, as an int, a non-negative pseudo-random number in (0,n)
     */
    export function rand(n: number): number

    export function abs(n: number): number

    export function min(nums: number[]): number

    export function floor(n: number): number
    export function ceil(n: number): number

    export function round(n: number, decimals?: number): number

    export function median(nums: number[]): number

    export function standardDev(nums: number[]): number
}





declare namespace net {
    export function inCIDR(cidr: string, ip: string): boolean;

    export function getIPAddress(): string

    export function getMacAddress(): string

    export type dialNetwork = "tcp" | "tcp4" | "tcp6" | "udp" | "udp4" | "udp6" | "ip" | "ip4" | "ip6" | "unix" | "unixgram" | "unixpacket"

    export type listenNetwork = "tcp" | "tcp4" | "tcp6" | "unix" | "unixpacket"

    export interface IP {
        string(): string
    }

    export interface Connection {
        read(b: byte[]): number
        write(b: byte[]): number
        setDeadline(t: time.Time): void
        setWriteDeadline(t: time.Time): void
        setReadDeadline(t: time.Time): void
        close(): void
    }

    export interface Listener {
        accept(): Connection
        close(): void
    }

    export function dial(network: dialNetwork, address: string): Connection
    export function dialTimeout(network: dialNetwork, address: string, d: time.Duration | number): Connection
    export function listen(network: listenNetwork, address: string): Listener

    export interface TCPListener {
        accept(): TCPConnection
        close(): void
    }

    export function dialTCP(network: dialNetwork, localAddr: TCPAddr, remoteAddr: TCPAddr): TCPConnection
    export function listenTCP(network: listenNetwork, address: TCPAddr): TCPListener

    export interface TCPConnection {
        localAddr: TCPAddr | Addr
        remoteAddr: TCPAddr | Addr
        read(b: byte[]): number
        write(b: byte[]): number
        setDeadline(t: time.Time): void
        setWriteDeadline(t: time.Time): void
        setReadDeadline(t: time.Time): void
        close(): void
    }

    export function resolveTCPAddr(network: dialNetwork, address: string): TCPAddr

    export interface TCPAddr {
        IP: IP
        port: number
        IPAddress(): string
        string(): string
    }

    export interface Addr {
        IPAddress(): string
        string(): string
    }
}





declare namespace os {
    export const stdin: io.File
    export const stdout: io.File
    export const stderr: io.File
    export const fileSystem: io.FileSystem

    export function readLine(): string

    export function exec(name: string, ...params: string[]): string

    /**
     * Reads an environment variable.
     */
    export function getEnv(key: string): string
    /**
     * Sets an environment variable.
     */
    export function setEnv(key: string, value: string): void

    export function exit(code?: number): void

    export const homeDir: string
    export const pathSeparator: string

    export function hostName(): string

    export function mapPath(path: string): string

    export function newUUID(): string

    export function newCommand(name: string, ...params: any[]): Command

    export interface Command {
        dir: string
        env: string[]
        stdin: io.File
        stdout: io.File
        stderr: io.File

        run(): void
        start(): void
        output(): string
        combinedOutput(): string
    }
}






declare namespace png {

    export function encode(w: io.Writer, img: Image): void

    export function decode(buf: byte[] | io.Reader): Image

    export interface Image { }
}






declare namespace reflect {
    export const program: runtime.Program

    export function is<T>(v: any, name: string): v is T

    export function typeOf(v: any): string

    export function isValue(v: any): boolean
    export function isNativeObject(v: any): boolean
    export function isArray(v: any): boolean
    export function isMap(v: any): boolean

    export function getFunction(name: string): Function

    export function call(name: string, ...params: any[]): any

    export function runFunc(name: string, ...params: any[]): any
}






declare namespace regex {
    export function match(pattern: string, value: string): boolean
    export function split(pattern: string, value: string): string[]
    export function findAllStringSubmatch(pattern: string, value: string, count?: number): string[]
    export function findAllStringSubmatchIndex(pattern: string, value: string, count?: number): number[][]
    export function replaceAllString(pattern: string, source: string, replace: string): string
}





declare namespace rsa {
    export function generateKey(size?: number): PrivateKey
    export function decodePEMKey(key: string | byte[]): PrivateKey
    export function decodePublicPEMKey(key: string | byte[]): PublicKey
    export function signPKCS1v15(key: PrivateKey, mesage: string | byte[]): byte[]
    export function verifyPKCS1v15(key: PublicKey, mesage: string | byte[], signature: string | byte[]): boolean

    interface PrivateKey {
        publicKey: PublicKey
        encodePEMKey(): byte[]
        encodePublicPEMKey(): byte[]
    }

    interface PublicKey {

    }
}





declare function panic(message: string): void
declare function defer(f: () => void): void;

declare namespace runtime {
    export const version: string
}

declare namespace runtime {
    export interface Finalizable {
        close(): void
    }

    export function typeDefs(): string

    export function setFileSystem(fs: io.FileSystem): void

    export function setFinalizer(v: runtime.Finalizable): void
    export function newFinalizable(f: () => void): Finalizable

    export function panic(message: string): void

    export type OSName = "linux" | "windows" | "darwin"

    export const pluginManager: PluginManager
    export const globalPluginManager: PluginManager

    export function setGlobalPluginManager(p: PluginManager): void
    export function initGlobalPluginManager(fs?: io.FileSystem): PluginManager

    export const context: Context

    // allows to set values shared accross all contexts
    export const global: any

    export function newPluginManager(fs?: io.FileSystem): PluginManager
    export function newPlugin(name: string, program: Program): Plugin
    export function newContext(): Context

    /**
     * Returns the operating system
     */
    export const OS: OSName

    /**
     * Returns the path of the executable.
     */
    export const executable: string

    /**
     * Returns the path of the native runtime executable.
     */
    export const nativeExecutable: string

    export const vm: VirtualMachine


    export function runFunc(func: string, ...args: any[]): any
    export function exec(func: string, ...args: any[]): any
    export function execIfExists(func: string, ...args: any[]): any

    // export function attachHook(name: string, func: Function): void
    // export function dispatchHook(name: string, ...params: any[]): void
    // export function anyHook(name: string): boolean

    export const hasResources: boolean
    export function resources(name: string): string[]
    export function resource(name: string): byte[]

    export function getStackTrace(): string
    export function newVM(p: Program, globals?: any[]): VirtualMachine

    export interface Program {
        functions(): FunctionInfo[]
        functionInfo(name: string): FunctionInfo
        resources(): string[]
        resource(key: string): byte[]
        setResource(key: string, value: byte[]): void

        attributes(): StringMap
        attribute(name: string): string
        attributeValues(name: string): string[]
        addAttribute(name: string, value: string): void

        /**
         * Strip sources, not exported functions name and other info.
         */
        strip(): void
        toString(): string
        write(w: io.Writer): void
    }

    export interface FunctionInfo {
        name: string
        index: number
        arguments: number
        exported: boolean
        func: Function
        toString(): string
    }

    export interface VirtualMachine {
        maxAllocations: number
        maxFrames: number
        maxSteps: number
        readonly steps: number
        readonly allocations: number
        readonly fileSystem: io.FileSystem
        readonly program: Program
        context: Context
        trusted: boolean
        error: errors.Error
        initialize(): void
        run(...args: any[]): any
        runFunc(name: string, ...args: any[]): any
        runFunc(index: number, ...args: any[]): any
        runStaticFunc(name: string, ...args: any[]): any
        runStaticFunc(index: number, ...args: any[]): any
        getValue(name: string): any
        getGlobals(): any[]
        getStackTrace(): string
        setFileSystem(v: io.FileSystem): void
        clone(): VirtualMachine
        resetSteps(): void
    }

    export interface Context {
        db: sql.DB
        plugin: Plugin
        pluginManager: PluginManager
        pluginName: string
        language: string
        culture: i18n.Culture
        location: time.Location
        debug: boolean
        test: boolean
        items: { [key: string]: any }
        dataFS: io.FileSystem
        addPlugin(name: string): void
        hasPlugin(name: string): boolean
        getPlugins(): string[]
        setPlugins(v: string[]): void
        clone(): Context
        exec(func: string, ...args: any[]): any
        execIfExists(func: string, ...args: any[]): any
        getProtectedItem(name: string): any
        setProtectedItem(name: string, v: any): void
    }

    export interface PluginManager {
        debug: boolean
        fileSystem: io.FileSystem
        getPlugin(plugin: string): Plugin
        allPlugins(): Plugin[]
        loadPlugin(plugin: string | Plugin, trusted?: boolean): Plugin
        reloadPlugin(plugin: string, trusted?: boolean): Plugin
        clone(): PluginManager
        clear(name?: string): void
        runFunc(func: string, ...args: any[]): any
        exec(c: Context, func: string, ...args: any[]): any
        execIfExists(c: Context, func: string, ...args: any[]): any
        copy(): PluginManager

        attachHook(name: string, func: Function): void
        dispatchHook(name: string, ...params: any[]): void
        anyHook(name: string): boolean
    }

    export interface Plugin {
        name: string
        dirName: string
        program: Program
        globals: any[]
        setGlobals(v: any[]): void
    }
}






declare namespace smtp {
    export function newMessage(): Message

    export function send(
        msg: Message,
        user: string,
        password: string,
        host: string,
        port: number,
        insecureSkipVerify?: boolean): void

    export interface Message {
        from: string
        fromName: string
        to: string[]
        cc: string[]
        bcc: string[]
        replyTo: string
        subject: string
        body: string
        html: boolean
        toString(): string
        attach(fileName: string, data: byte[], inline: boolean): void
    }
}







declare namespace sql {
    export type DriverType = "mysql"

    /**
     * If you specify a databaseName every query will be parsed and all tables will be
     * prefixed with the database name: "SELECT foo FROM bar" will automatically be converted 
     * to "SELECT databasename.foo FROM bar". 
     * 
     * The parser understands most SQL standard but very little DDL (no ALTER TABLE yet).
     */
    export function open(driver: DriverType, connString: string, databaseName?: string): DB
    export function changeDatabase(name: string): void

    export function setLogAllQueries(value: boolean): void
    //export function setAuditAllQueries(value: boolean): void

    export let hasTransaction: boolean
    export let nestedTransactions: number
    export let driver: DriverType
    export let database: string
    export let transactionNestLevel: number

    export function exec(query: string | Query, ...params: any[]): Result
    export function execRaw(query: string, ...params: any[]): Result
    export function reader(query: string | SelectQuery, ...params: any[]): Reader
    export function query(query: string | SelectQuery, ...params: any[]): any[]
    export function queryValues(query: string | SelectQuery, ...params: any[]): any[]
    export function queryRaw(query: string | SelectQuery, ...params: any[]): any[]
    export function queryFirst(query: string | SelectQuery, ...params: any[]): any
    export function queryValue(query: string | SelectQuery, ...params: any[]): any

    export function queryValueRaw(query: string | SelectQuery, ...params: any[]): any
    export function loadTable(query: string | SelectQuery, ...params: any[]): Table
    export function beginTransaction(): void
    export function commit(force?: boolean): void
    export function rollback(): void
    export function hasTable(name: string): boolean
    export function databases(): string[]
    export function tables(): string[]
    export function columns(table: string): SchemaColumn[]
    export function setWhitelistFuncs(funcs: string[]): void

    /**
     * DB is a handle to the database.
     */
    export interface DB {
        database: string
        namespace: string
        openAnyDatabase: boolean
        readOnly: boolean
        driver: DriverType
        nestedTransactions: number
        hasTransaction: boolean

        setMaxOpenConns(v: number): void
        setMaxIdleConns(v: number): void
        setConnMaxLifetime(d: time.Duration | number): void

        onQuery: (query: string | SelectQuery, ...params: any[]) => void
        open(name: string): DB
        close(): void

        reader(query: string | SelectQuery, ...params: any[]): Reader
        query(query: string | SelectQuery, ...params: any[]): any[]
        queryRaw(query: string | SelectQuery, ...params: any[]): any[]
        queryFirst(query: string | SelectQuery, ...params: any[]): any
        queryFirstRaw(query: string | SelectQuery, ...params: any[]): any
        queryValues(query: string | SelectQuery, ...params: any[]): any[]
        queryValuesRaw(query: string | SelectQuery, ...params: any[]): any[]
        queryValue(query: string | SelectQuery, ...params: any[]): any
        queryValueRaw(query: string | SelectQuery, ...params: any[]): any

        loadTable(query: string | SelectQuery, ...params: any[]): Table
        loadTableRaw(query: string | SelectQuery, ...params: any[]): Table

        exec(query: string | Query, ...params: any[]): Result
        execRaw(query: string, ...params: any[]): Result

        beginTransaction(): void
        commit(): void
        rollback(): void

        hasDatabase(name: string): boolean
        hasTable(name: string): boolean
        databases(): string[]
        tables(): string[]
        columns(table: string): SchemaColumn[]
    }

    export interface SchemaColumn {
        name: string
        type: string
        size: number
        decimals: number
        nullable: boolean
    }

    export interface Reader {
        next(): boolean
        read(): any
        readValues(): any[]
        close(): void
    }

    export interface Result {
        lastInsertId: number
        rowsAffected: number
    }

    export interface Table {
        columns: Column[]
        rows: Row[]
        length: number
        page?: number
        pageSize?: number
        totalCount?: number
        items?: any
    }

    export interface Row extends Array<any> {
        [index: number]: any
        [key: string]: any
        length: number
        columns: Array<Column>
    }

    export type ColumnType = "string" | "int" | "float" | "bool" | "datetime"

    export interface Column {
        name: string
        type: ColumnType
    }

    export function parse(query: string, ...params: any[]): Query
    export function select(query: string, ...params: any[]): SelectQuery

    export interface ValidateOptions {
        tables: Map<string[]>
    }

    export function validateSelect(q: SelectQuery, options: ValidateOptions): void

    export function newSelect(): SelectQuery

    export function where(filter: string, ...params: any[]): SelectQuery

    export function orderBy(s: string): SelectQuery

    export interface Query {
        toSQL(format?: boolean, driver?: DriverType, escapeIdents?: boolean, ignoreNamespaces?: boolean): string
    }

    export interface DQLQuery extends Query {
        hasLimit: boolean
        hasWhere: boolean
        parameters: any[]
        where(s: string, ...params: any[]): SelectQuery
        and(s: string, ...params: any[]): SelectQuery
        and(filter: SelectQuery): SelectQuery
        or(s: string, ...params: any[]): SelectQuery
        limit(rowCount: number): SelectQuery
        limit(rowCount: number, offset: number): SelectQuery
    }

    export interface SelectQuery extends Query {
        columnsLength: number
        hasLimit: boolean
        hasFrom: boolean
        hasWhere: boolean
        hasDistinct: boolean
        hasOrderBy: boolean
        hasUnion: boolean
        hasGroupBy: boolean
        hasHaving: boolean
        parameters: any[]
        addColumns(s: string): SelectQuery
        setColumns(s: string): SelectQuery
        from(s: string): SelectQuery
        fromExpr(q: SelectQuery, alias: string): SelectQuery
        limit(rowCount: number): SelectQuery
        limit(rowCount: number, offset: number): SelectQuery
        groupBy(s: string): SelectQuery
        orderBy(s: string): SelectQuery
        where(s: string, ...params: any[]): SelectQuery
        having(s: string, ...params: any[]): SelectQuery
        and(s: string, ...params: any[]): SelectQuery
        and(filter: SelectQuery): SelectQuery
        or(s: string, ...params: any[]): SelectQuery
        join(s: string, ...params: any[]): SelectQuery

        /**
         * copies all the elements of the query from the Where part.
         */
        setFilter(q: SelectQuery): void

        getFilterColumns(): string[]
    }

    export interface UpdateQuery extends Query {
        hasLimit: boolean
        hasWhere: boolean
        parameters: any[]
        addColumns(s: string, ...params: any[]): UpdateQuery
        setColumns(s: string, ...params: any[]): UpdateQuery
        where(s: string, ...params: any[]): UpdateQuery
        and(s: string, ...params: any[]): UpdateQuery
        and(filter: UpdateQuery): UpdateQuery
        or(s: string, ...params: any[]): UpdateQuery
        limit(rowCount: number): UpdateQuery
        limit(rowCount: number, offset: number): UpdateQuery
    }

    export interface DeleteQuery extends Query {
        hasLimit: boolean
        hasWhere: boolean
        parameters: any[]
        where(s: string, ...params: any[]): DeleteQuery
        and(s: string, ...params: any[]): DeleteQuery
        and(filter: DeleteQuery): DeleteQuery
        or(s: string, ...params: any[]): DeleteQuery
        limit(rowCount: number): DeleteQuery
        limit(rowCount: number, offset: number): DeleteQuery
    }
}






declare namespace strconv {
    export function formatInt(i: number, base: number): string
    export function parseInt(s: string, base: number, bitSize: number): number
    export function formatCustomBase34(i: number): string
    export function parseCustomBase34(s: string): number

}





declare namespace strings {
    export function newReader(a: string): io.Reader
}

interface String {
    runeAt(i: number): string
}

declare namespace strings {
    export function equalFold(a: string, b: string): boolean
    export function isChar(value: string): boolean
    export function isDigit(value: string): boolean
    export function isIdent(value: string): boolean
    export function isAlphanumeric(value: string): boolean
    export function isAlphanumericIdent(value: string): boolean
    export function isNumeric(value: string): boolean
    export function sort(a: string[]): void
}

interface String {
    [n: number]: string

    /**
     * Gets the length of the string.
     */
    length: number

    /**
     * The number of bytes oposed to the number of runes returned by length.
     */
    runeCount: number

    toLower(): string

    toUpper(): string

    toTitle(): string

    toUntitle(): string

    replace(oldValue: string, newValue: string, times?: number): string

    hasPrefix(prefix: string): boolean
    hasSuffix(prefix: string): boolean

    trim(cutset?: string): string
    trimLeft(cutset?: string): string
    trimRight(cutset?: string): string
    trimPrefix(prefix: string): string
    trimSuffix(suffix: string): string

    rightPad(pad: string, total: number): string
    leftPad(pad: string, total: number): string

    take(to: number): string
    substring(from: number, to?: number): string
    runeSubstring(from: number, to?: number): string

    split(s: string): string[]
    splitEx(s: string): string[]

    contains(s: string): boolean
    equalFold(s: string): boolean

    indexOf(s: string, start?: number): number
    lastIndexOf(s: string, start?: number): number


    /**
     * Replace with regular expression.
     * The syntax is defined: https://golang.org/pkg/regexp/syntax
     */
    replaceRegex(expr: string, replace: string): string
}





declare namespace sync {
    export function newMutex(): Mutex
    export function newWaitGroup(concurrency?: number): WaitGroup

    export function execLocked(key: string, func: Function): any

    export interface WaitGroup {
        go(f: Function): void
        wait(): void
    }

    export interface Mutex {
        lock(): void
        unlock(): void
    }

    export function newChannel(buffer?: number): Channel

    export function select(channels: Channel[], defaultCase?: boolean): { index: number, value: any, receivedOK: boolean }

    export interface Channel {
        send(v: any): void
        receive(): any
        close(): void
    }
}






declare namespace templates {
    /**
     * Reads the file and processes includes
     */
    export function exec(code: string, model?: any): string
    export function preprocess(path: string, fs?: io.FileSystem): string
    export function render(text: string, model?: any): string
    export function renderHTML(text: string, model?: any): string
    /**
     * 
     * @param headerFunc By defauult is: function render(w: io.Writer, model: any)
     */
    export function compile(text: string, headerFunc?: string): string
    /**
     * 
     * @param headerFunc By defauult is: function render(w: io.Writer, model: any)
     */
    export function compileHTML(text: string, headerFunc?: string): string

    /**
     * 
     * @param headerFunc By defauult is: function render(w: io.Writer, model: any)
     */
    export function writeHTML(w: io.Writer, path: string, model?: any, fs?: io.FileSystem, headerFunc?: string): void

    /**
     * 
     * @param headerFunc By defauult is: function render(w: io.Writer, model: any)
     */
    export function writeHTMLTemplate(w: io.Writer, template: string, model?: any, headerFunc?: string): void
}





declare namespace time {
    /**
     * The ISO time format.
     */
    export const RFC3339: string
    /**
     * The default date format.
     */
    export const DefaultDateFormat: string

    export const AllWeekDays: number

    export const Nanosecond: number
    export const Microsecond: number
    export const Millisecond: number
    export const Second: number
    export const Minute: number
    export const Hour: number

    export const SecMillis: number
    export const MinMillis: number
    export const HourMillis: number
    export const DayMillis: number

    export function now(): Time
    export function nowUTC(): Time

    export const Monday: number
    export const Tuesday: number
    export const Wednesday: number
    export const Thursday: number
    export const Friday: number
    export const Saturday: number
    export const Sunday: number

    /**
     * The number of nanoseconds since the unix epoch.
     */
    export let unixNano: number

    export interface Location {
        name: string
    }

    export const utc: Location
    export const local: Location

    export function setDefaultLocation(name: string): void

    /**
     * Sets a fixed value for now() for testing.
     */
    export function setFixedNow(t: time.Time): void

    /**
     * Remove a fixed value for now().
     */
    export function unsetFixedNow(): void
    export function loadLocation(name: string): Location

    export function convertFormat(format: string): string
    export function formatMinutes(v: number): string

    export function setDayOfWeek(value: number, dayOfWeek: number, active: boolean): number
    export function isDayOfWeekActive(value: number, dayOfWeek: number): boolean

    /**
     * 
     * @param seconds from unix epoch
     */
    export function unix(seconds: number): Time

    export function date(year?: number, month?: number, day?: number, hour?: number, min?: number, sec?: number, loc?: Location): Time
    export function localDate(year?: number, month?: number, day?: number, hour?: number, min?: number, sec?: number): Time


    export function parseDuration(s: string): Duration

    export function duration(nanoseconds: number | Duration): Duration
    export function toDuration(hour: number, minute?: number, second?: number): Duration
    export function toMilliseconds(hour: number, minute?: number, second?: number): number

    export function daysInMonth(year: number, month: number): number

    export interface Time {
        unix: number
        second: number
        nanosecond: number
        minute: number
        hour: number
        day: number
        /**
         * sunday = 0, monday = 1, ...
         */
        dayOfWeek: number
        month: number
        year: number
        yearDay: number
        location: Location
        /**
         * The time part in milliseconds
         */
        time(): number

        /**
         * Return the date discarding the time part in local time.
         */
        startOfDay(): Time
        /**
         * Returns the las moment of the day in local time
         */
        endOfDay(): Time
        utc(): Time
        local(): Time
        sub(t: Time): Duration
        add(t: Duration | number): Time
        addYears(t: number): Time
        addMonths(t: number): Time
        addDays(t: number): Time
        addHours(t: number): Time
        addMinutes(t: number): Time
        addSeconds(t: number): Time
        addMilliseconds(t: number): Time

        setDate(year?: number, month?: number, day?: number): Time
        setTime(hour?: number, minute?: number, second?: number, millisecond?: number): Time
        setTimeMillis(millis: number): Time

        format(f: string): string
        formatIn(f: string, loc: Location): string
        in(loc: Location): Time
        /**
         * setLocation returns the same time with the location. No conversions
         * are made. 9:00 UTC becomes 9:00 Europe/Madrid
         */
        setLocation(loc: Location): Time
        equal(t: Time): boolean
        after(t: Time): boolean
        afterOrEqual(t: Time): boolean
        before(t: Time): boolean
        beforeOrEqual(t: Time): boolean
        between(t1: Time, t2: Time): boolean
        sameDay(t: Time): boolean
    }

    export interface Duration {
        hours: number
        minutes: number
        seconds: number
        milliseconds: number
        nanoseconds: number
        equal(other: number | Duration): boolean
        greater(other: number | Duration): boolean
        lesser(other: number | Duration): boolean
        add(other: number | Duration): Duration
        sub(other: number | Duration): Duration
        multiply(other: number | Duration): Duration
    }

    export interface Period {
        start?: time.Time
        end?: time.Time
    }

    export function after(d: number | Duration, value?: any): sync.Channel
    export function sleep(millis: number): void
    export function sleep(d: time.Duration): void
    export function parse(value: any, format?: string): time.Time
    export function parseLocal(value: any, format?: string): time.Time
    export function parseInLocation(value: any, format: string, location: Location): time.Time


    export function newTicker(duration: number | time.Duration, func: Function): Ticker
    export function newTimer(duration: number, func: Function): Ticker

    export interface Ticker {
        stop(): void
    }

}







declare namespace tls {
    export function newConfig(insecureSkipVerify?: boolean): Config

    export interface Config {
        insecureSkipVerify: boolean
        loadCertificate(certPath: string, keyPath: string): void
        loadCertificateData(cert: byte[] | string, key: byte[] | string): void
    }

    export interface Certificate {
        cert: byte[]
        key: byte[]
    }

    export function generateCert(): Certificate
}





declare namespace websocket {
    export function upgrade(r: http.Request): WebsocketConnection

    export interface WebsocketConnection {
        guid: string
        write(v: any): number | void
        writeJSON(v: any): void
        writeText(text: string | byte[]): void
        readMessage(): WebSocketMessage
        close(): void
    }

    export interface WebSocketMessage {
        type: WebsocketType
        message: string
    }

    export enum WebsocketType {
        text = 1,
        binary = 2,
        close = 8,
        ping = 9,
        pong = 10
    }
}





declare namespace sync {
    export function newWorker(): Worker
    export function newJob(): Job

    export interface Worker {
        readonly isRunning: boolean
        errorFunc: (job: any, e: errors.Error) => void
        add(job: Job): void
        start(): void
        stop(): void
    }

    export interface Job {
        idTask?: number
        params: any
        timeout?: time.Duration
        workFunc: (params: any) => void
    }
}





declare namespace xlsx {
    export function newFile(): XLSXFile
    export function openFile(path: string): XLSXFile
    export function openFile(file: io.File): XLSXFile
    export function openReaderAt(r: io.ReaderAt, size: number): XLSXFile
    export function openBinary(file: io.File): XLSXFile
    export function newStyle(): Style

    export interface XLSXFile {
        sheets: XLSXSheet[]
        addSheet(name: string): XLSXSheet
        save(path?: string): void
        write(w: io.Writer): void
    }

    export interface XLSXSheet {
        rows: XLSXRow[]
        col(i: number): Col
        addRow(): XLSXRow
    }

    export interface Col {
        width: number
    }

    export interface XLSXRow {
        cells: XLSXCell[]
        height: number
        addCell(v?: any): XLSXCell
    }

    export interface XLSXCell {
        value: any
        numberFormat: string
        style: Style
        getDate(): time.Time
        merge(hCells: number, vCells: number): void
    }

    export interface Style {
        alignment: Alignment
        applyAlignment: boolean
        font: Font
        applyFont: boolean
    }

    export interface Alignment {
        horizontal: string
        vertical: string
    }

    export interface Font {
        bold: boolean
        size: number
    }
}





declare namespace xml {
    export function newDocument(): XMLDocument

    export function readString(s: string): XMLDocument

    export interface XMLDocument {
        createElement(name: string): XMLElement
        selectElement(name: string): XMLElement
        toString(): string
    }

    export interface XMLElement {
        tag: string
        selectElements(name: string): XMLElement[]
        selectElement(name: string): XMLElement
        createElement(name: string): XMLElement
        createAttribute(name: string, value: string): XMLElement
        getAttribute(name: string): string
        setValue(value: string | number | boolean): void
        getValue(): string
    }
}







declare namespace zip {
    export function newWriter(w: io.Writer): Writer
    export function newReader(r: io.Reader, size: number): io.ReaderCloser
    export function open(path: string, fs?: io.FileSystem): Reader

    export interface Writer {
        create(name: string): io.Writer
        flush(): void
        close(): void
    }

    export interface Reader {
        files(): File[]
        close(): void
    }

    export interface File {
        name: string
        compressedSize: number
        uncompressedSize: number
        open(): io.ReaderCloser
    }
}


