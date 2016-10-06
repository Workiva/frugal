package com.workiva.frugal.transport;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.ObjectWriter;
import com.fasterxml.jackson.databind.ObjectMapper.DefaultTyping;

/**
 * This class implements conversion through Jackson JSON processor. Methods are
 * provided for serializing an object to JSON and deserializing from JSON to an
 * object.
 */
public class JsonDataConverter {

	protected final ObjectMapper objectMapper;

	public JsonDataConverter() {
		this(new ObjectMapper());
		objectMapper.configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
		objectMapper.configure(SerializationFeature.FAIL_ON_EMPTY_BEANS, false);
		objectMapper.enableDefaultTyping(DefaultTyping.NON_FINAL);
	}

	public JsonDataConverter(ObjectMapper objectMapper) {
		this.objectMapper = objectMapper;
	}

	public String serializeToJson(Object obj) throws JsonProcessingException {
		ObjectWriter objectWriter = objectMapper.writer();
		return objectWriter.writeValueAsString(obj);
	}

	public <T> T deserializeFromJson(String jsonText, Class<T> objectType) throws Exception {
		return objectMapper.readValue(jsonText, objectType);
	}

}