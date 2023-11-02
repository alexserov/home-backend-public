import ModbusRTU from "modbus-serial";

export type ModbusAction<T = void> = (client: ModbusRTU)=>Promise<T>;
