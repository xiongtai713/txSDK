/*************************************************************************
 * Copyright (C) 2016-2019 PDX Technologies, Inc. All Rights Reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 *************************************************************************/

package ltd.pdx.utopia.sample.sample;

import ltd.pdx.utopia.shim.ChaincodeBaseX;
import org.hyperledger.fabric.shim.ChaincodeStub;
import org.hyperledger.fabric.shim.ledger.KeyModification;
import org.hyperledger.fabric.shim.ledger.KeyValue;
import org.hyperledger.fabric.shim.ledger.QueryResultsIterator;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.InputStream;
import java.util.Arrays;
import java.util.List;
import java.util.Map;

public class MyCc extends ChaincodeBaseX {

private static Logger logger = LoggerFactory.getLogger(MyCc.class);

@Override
public Response init(ChaincodeStub stub) {
logger.info("init............................");
return newSuccessResponse();
}

@Override
public Response invoke(ChaincodeStub stub) {
logger.info("invoke..........................");
String payload = "";
try {
final String function = stub.getFunction();
final List<String> params = stub.getParameters();

System.out.println("invoke method receive function :" + function);
System.out.println("invoke method receive params :" + params);
switch (function) {
case "getHis":
String hisResult = "";
QueryResultsIterator<KeyModification> historyForKey = stub.getHistoryForKey(params.get(0));
for (KeyModification e : historyForKey) {
hisResult += String.format("tid : %s ;key : %s ;value : %s ;isDelete : %s \n", e.getTxId(), params.get(0), e.getStringValue(), e.isDeleted());
}

historyForKey.close();
payload = hisResult;
break;
case "get":
String state = stub.getStringState(params.get(0));
logger.info("get key : {} result : {}", params.get(0), state);
payload = state;
break;
case "put":
stub.putStringState(params.get(0), params.get(1));
logger.info("put key : {} value : {}", params.get(0), params.get(1));
break;
case "del":
stub.delState(params.get(0));
logger.info("delete key : {}", params.get(0));
break;
case "getRange":
String rangeResult = "";
QueryResultsIterator<KeyValue> stateByRange = stub.getStateByRange(params.get(0), params.get(1));
for (KeyValue e : stateByRange) {
rangeResult += String.format("key : %s ;value : %s \n", e.getKey(), e.getStringValue());
}
//stateByRange.forEach(e -> logger.info("key : {} /value : {}", e.getKey(), e.getStringValue()));
stateByRange.close();
payload = rangeResult;
break;
case "stream":
Map<String, InputStream> streamMap = stub.getStream();
if (streamMap == null) {
logger.error("can not get stream!");
}
logger.info("stream size : {}", streamMap.size());
for (String name : streamMap.keySet()) {
logger.info("stream name : {}", name);
}
break;
}
} catch (Exception e) {
e.printStackTrace();
return newErrorResponse(e);
}

return newSuccessResponse(payload.getBytes());
}

public static void main(String[] args) throws Exception {

args = new String[]{"-a", "127.0.0.1:9052", "-i", "mycc:v1.0", "-c", "739"};
Arrays.stream(args).forEach(System.out::println);
MyCc myCc = new MyCc();
myCc.start(args);
}
}
